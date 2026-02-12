package httpapi

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/dimas1q/dockslim/backend/internal/apitokens"
	"github.com/dimas1q/dockslim/backend/internal/auth"
	"github.com/dimas1q/dockslim/backend/internal/dashboard"
	"github.com/dimas1q/dockslim/backend/internal/featureflags"
	"github.com/google/uuid"
)

type AccountHandler struct {
	authService               *auth.Service
	tokenService              APITokenService
	subscriptionService       SubscriptionService
	dashboardService          AccountDashboardService
	internalSubscriptionToken string
}

type APITokenService interface {
	CreateToken(ctx context.Context, userID uuid.UUID, name string, expiresAt *time.Time) (apitokens.Token, string, error)
	ListTokens(ctx context.Context, userID uuid.UUID) ([]apitokens.Token, error)
	RevokeToken(ctx context.Context, userID, tokenID uuid.UUID) error
}

type SubscriptionService interface {
	GetUserFeatures(ctx context.Context, userID uuid.UUID) (featureflags.UserFeatures, error)
	UpdateUserSubscription(ctx context.Context, input featureflags.UpdateSubscriptionInput) (featureflags.UserFeatures, error)
}

type AccountDashboardService interface {
	GetDashboard(ctx context.Context, userID uuid.UUID) (dashboard.AccountDashboard, error)
}

type AccountHandlerOptions struct {
	SubscriptionService       SubscriptionService
	DashboardService          AccountDashboardService
	InternalSubscriptionToken string
}

func NewAccountHandler(authService *auth.Service, tokenService APITokenService, options ...AccountHandlerOptions) *AccountHandler {
	handler := &AccountHandler{
		authService:  authService,
		tokenService: tokenService,
	}
	if len(options) > 0 {
		handler.subscriptionService = options[0].SubscriptionService
		handler.dashboardService = options[0].DashboardService
		handler.internalSubscriptionToken = options[0].InternalSubscriptionToken
	}
	return handler
}

type accountResponse struct {
	ID    string `json:"id"`
	Login string `json:"login"`
	Email string `json:"email"`
}

func (h *AccountHandler) Me(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	writeJSON(w, http.StatusOK, accountResponse{
		ID:    user.ID.String(),
		Login: user.Login,
		Email: user.Email,
	})
}

type updateProfileRequest struct {
	Login *string `json:"login,omitempty"`
	Email *string `json:"email,omitempty"`
}

func (h *AccountHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req updateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Login == nil && req.Email == nil {
		writeError(w, http.StatusBadRequest, "no fields to update")
		return
	}

	params := auth.UpdateProfileParams{
		Login: req.Login,
		Email: req.Email,
	}

	updated, err := h.authService.UpdateProfile(r.Context(), user.ID.String(), params)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrInvalidEmail), errors.Is(err, auth.ErrInvalidLogin):
			writeError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, auth.ErrEmailAlreadyExists), errors.Is(err, auth.ErrLoginAlreadyExists):
			writeError(w, http.StatusConflict, err.Error())
		case errors.Is(err, auth.ErrUserNotFound):
			writeError(w, http.StatusNotFound, "user not found")
		default:
			writeError(w, http.StatusInternalServerError, "failed to update profile")
		}
		return
	}

	writeJSON(w, http.StatusOK, accountResponse{
		ID:    updated.ID.String(),
		Login: updated.Login,
		Email: updated.Email,
	})
}

type createAPITokenRequest struct {
	Name      string  `json:"name"`
	ExpiresAt *string `json:"expires_at,omitempty"`
}

type createAPITokenResponse struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	Token     string     `json:"token"`
}

type apiTokenResponse struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
}

func (h *AccountHandler) CreateAPIToken(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req createAPITokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var expiresAt *time.Time
	if req.ExpiresAt != nil && *req.ExpiresAt != "" {
		parsed, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid expires_at format")
			return
		}
		expiresAt = &parsed
	}

	token, plain, err := h.tokenService.CreateToken(r.Context(), user.ID, req.Name, expiresAt)
	if err != nil {
		switch {
		case errors.Is(err, apitokens.ErrInvalidName):
			writeError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, apitokens.ErrNameConflict):
			writeError(w, http.StatusConflict, "token name already exists")
		default:
			writeError(w, http.StatusInternalServerError, "failed to create api token")
		}
		return
	}

	resp := createAPITokenResponse{
		ID:        token.ID.String(),
		Name:      token.Name,
		CreatedAt: token.CreatedAt,
		ExpiresAt: token.ExpiresAt,
		Token:     plain,
	}
	writeJSON(w, http.StatusCreated, resp)
}

func (h *AccountHandler) ListAPITokens(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	tokens, err := h.tokenService.ListTokens(r.Context(), user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list api tokens")
		return
	}

	resp := make([]apiTokenResponse, 0, len(tokens))
	for _, t := range tokens {
		resp = append(resp, apiTokenResponse{
			ID:         t.ID.String(),
			Name:       t.Name,
			LastUsedAt: t.LastUsedAt,
			CreatedAt:  t.CreatedAt,
			RevokedAt:  t.RevokedAt,
			ExpiresAt:  t.ExpiresAt,
		})
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *AccountHandler) RevokeAPIToken(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	tokenID, err := parseUUIDParam(r, "tokenId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid token id")
		return
	}

	if err := h.tokenService.RevokeToken(r.Context(), user.ID, tokenID); err != nil {
		if errors.Is(err, apitokens.ErrTokenNotFound) {
			writeError(w, http.StatusNotFound, "token not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to revoke token")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type accountSubscriptionPlanResponse struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Status     string     `json:"status"`
	ValidUntil *time.Time `json:"valid_until,omitempty"`
	IsAdmin    bool       `json:"is_admin"`
}

type accountSubscriptionResponse struct {
	Plan     accountSubscriptionPlanResponse `json:"plan"`
	Features map[string]any                  `json:"features"`
	Limits   map[string]any                  `json:"limits"`
}

func (h *AccountHandler) GetSubscription(w http.ResponseWriter, r *http.Request) {
	if h.subscriptionService == nil {
		writeError(w, http.StatusNotFound, "subscription service is not configured")
		return
	}

	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	featureSet, err := h.subscriptionService.GetUserFeatures(r.Context(), user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch subscription")
		return
	}

	writeJSON(w, http.StatusOK, toAccountSubscriptionResponse(featureSet))
}

func (h *AccountHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	if h.dashboardService == nil {
		writeError(w, http.StatusNotFound, "dashboard service is not configured")
		return
	}

	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	data, err := h.dashboardService.GetDashboard(r.Context(), user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch dashboard")
		return
	}

	writeJSON(w, http.StatusOK, data)
}

type updateSubscriptionRequest struct {
	UserID     string  `json:"user_id"`
	PlanID     string  `json:"plan_id"`
	Status     string  `json:"status"`
	ValidUntil *string `json:"valid_until,omitempty"`
}

func (h *AccountHandler) UpdateSubscription(w http.ResponseWriter, r *http.Request) {
	if h.subscriptionService == nil {
		writeError(w, http.StatusNotFound, "subscription service is not configured")
		return
	}

	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if !user.IsAdmin {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	internalToken := strings.TrimSpace(r.Header.Get("X-DockSlim-Internal-Token"))
	if h.internalSubscriptionToken == "" || internalToken == "" || subtle.ConstantTimeCompare([]byte(internalToken), []byte(h.internalSubscriptionToken)) != 1 {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	var req updateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	targetUserID, err := uuid.Parse(strings.TrimSpace(req.UserID))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user_id")
		return
	}

	var validUntil *time.Time
	if req.ValidUntil != nil && strings.TrimSpace(*req.ValidUntil) != "" {
		parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(*req.ValidUntil))
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid valid_until format")
			return
		}
		validUntil = &parsed
	}

	result, err := h.subscriptionService.UpdateUserSubscription(r.Context(), featureflags.UpdateSubscriptionInput{
		UserID:     targetUserID,
		PlanID:     req.PlanID,
		Status:     req.Status,
		ValidUntil: validUntil,
	})
	if err != nil {
		switch {
		case errors.Is(err, featureflags.ErrPlanNotFound):
			writeError(w, http.StatusBadRequest, "invalid plan_id")
		default:
			writeError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	writeJSON(w, http.StatusOK, toAccountSubscriptionResponse(result))
}

func toAccountSubscriptionResponse(featureSet featureflags.UserFeatures) accountSubscriptionResponse {
	return accountSubscriptionResponse{
		Plan: accountSubscriptionPlanResponse{
			ID:         featureSet.PlanID,
			Name:       featureSet.PlanName,
			Status:     featureSet.Status,
			ValidUntil: featureSet.ValidUntil,
			IsAdmin:    featureSet.IsAdmin,
		},
		Features: featureSet.Features,
		Limits:   featureSet.Limits(),
	}
}

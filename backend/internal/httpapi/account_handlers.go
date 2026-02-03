package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/dimas1q/dockslim/backend/internal/apitokens"
	"github.com/dimas1q/dockslim/backend/internal/auth"
	"github.com/google/uuid"
)

type AccountHandler struct {
	authService  *auth.Service
	tokenService APITokenService
}

type APITokenService interface {
	CreateToken(ctx context.Context, userID uuid.UUID, name string, expiresAt *time.Time) (apitokens.Token, string, error)
	ListTokens(ctx context.Context, userID uuid.UUID) ([]apitokens.Token, error)
	RevokeToken(ctx context.Context, userID, tokenID uuid.UUID) error
}

func NewAccountHandler(authService *auth.Service, tokenService APITokenService) *AccountHandler {
	return &AccountHandler{authService: authService, tokenService: tokenService}
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

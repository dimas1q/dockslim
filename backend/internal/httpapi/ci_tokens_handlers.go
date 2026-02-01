package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/dimas1q/dockslim/backend/internal/auth"
	"github.com/dimas1q/dockslim/backend/internal/citokens"
	"github.com/google/uuid"
)

type CITokensHandler struct {
	service CITokenService
}

type CITokenService interface {
	CreateToken(ctx context.Context, userID, projectID uuid.UUID, name string, expiresAt *time.Time) (citokens.Token, string, error)
	ListTokens(ctx context.Context, userID, projectID uuid.UUID) ([]citokens.Token, error)
	RevokeToken(ctx context.Context, userID, projectID, tokenID uuid.UUID) error
}

func NewCITokensHandler(service CITokenService) *CITokensHandler {
	return &CITokensHandler{service: service}
}

type createCITokenRequest struct {
	Name      string  `json:"name"`
	ExpiresAt *string `json:"expires_at,omitempty"`
}

type createCITokenResponse struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	Token     string     `json:"token"`
}

type ciTokenResponse struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
}

func (h *CITokensHandler) Create(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	projectID, err := parseUUIDParam(r, "projectId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	var req createCITokenRequest
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

	token, plain, err := h.service.CreateToken(r.Context(), user.ID, projectID, req.Name, expiresAt)
	if err != nil {
		switch {
		case errors.Is(err, citokens.ErrInvalidName):
			writeError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, citokens.ErrProjectNotFound):
			writeError(w, http.StatusNotFound, "project not found")
		case errors.Is(err, citokens.ErrNotOwner):
			writeError(w, http.StatusForbidden, "forbidden")
		case errors.Is(err, citokens.ErrNameConflict):
			writeError(w, http.StatusConflict, "token name already exists")
		default:
			writeError(w, http.StatusInternalServerError, "failed to create ci token")
		}
		return
	}

	resp := createCITokenResponse{
		ID:        token.ID.String(),
		Name:      token.Name,
		CreatedAt: token.CreatedAt,
		ExpiresAt: token.ExpiresAt,
		Token:     plain,
	}
	writeJSON(w, http.StatusCreated, resp)
}

func (h *CITokensHandler) List(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	projectID, err := parseUUIDParam(r, "projectId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	tokens, err := h.service.ListTokens(r.Context(), user.ID, projectID)
	if err != nil {
		switch {
		case errors.Is(err, citokens.ErrProjectNotFound):
			writeError(w, http.StatusNotFound, "project not found")
		case errors.Is(err, citokens.ErrNotOwner):
			writeError(w, http.StatusForbidden, "forbidden")
		default:
			writeError(w, http.StatusInternalServerError, "failed to list ci tokens")
		}
		return
	}

	resp := make([]ciTokenResponse, 0, len(tokens))
	for _, t := range tokens {
		resp = append(resp, ciTokenResponse{
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

func (h *CITokensHandler) Revoke(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	projectID, err := parseUUIDParam(r, "projectId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}
	tokenID, err := parseUUIDParam(r, "tokenId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid token id")
		return
	}

	if err := h.service.RevokeToken(r.Context(), user.ID, projectID, tokenID); err != nil {
		switch {
		case errors.Is(err, citokens.ErrProjectNotFound):
			writeError(w, http.StatusNotFound, "project not found")
		case errors.Is(err, citokens.ErrNotOwner):
			writeError(w, http.StatusForbidden, "forbidden")
		case errors.Is(err, citokens.ErrTokenNotFound):
			writeError(w, http.StatusNotFound, "token not found")
		default:
			writeError(w, http.StatusInternalServerError, "failed to revoke token")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

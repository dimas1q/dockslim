package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/dimas1q/dockslim/backend/internal/auth"
)

type AuthHandler struct {
	service        *auth.Service
	accessTokenTTL time.Duration
}

func NewAuthHandler(service *auth.Service, accessTokenTTL time.Duration) *AuthHandler {
	return &AuthHandler{service: service, accessTokenTTL: accessTokenTTL}
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type registerResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.service.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrInvalidEmail), errors.Is(err, auth.ErrInvalidPassword):
			writeError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, auth.ErrEmailAlreadyExists):
			writeError(w, http.StatusConflict, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "failed to create user")
		}
		return
	}

	resp := registerResponse{ID: user.ID.String(), Email: user.Email}
	writeJSON(w, http.StatusCreated, resp)
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	_, token, err := h.service.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to login")
		return
	}

	resp := loginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int64(h.accessTokenTTL.Seconds()),
	}
	writeJSON(w, http.StatusOK, resp)
}

type meResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	resp := meResponse{ID: user.ID.String(), Email: user.Email}
	writeJSON(w, http.StatusOK, resp)
}

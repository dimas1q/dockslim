package httpapi

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/dimas1q/dockslim/backend/internal/auth"
)

type AuthHandler struct {
	service        *auth.Service
	accessTokenTTL time.Duration
	cookieConfig   CookieConfig
}

type CookieConfig struct {
	SameSite http.SameSite
	Secure   bool
	Domain   string
	Path     string
}

func NewAuthHandler(service *auth.Service, accessTokenTTL time.Duration, cookieConfig CookieConfig) *AuthHandler {
	return &AuthHandler{service: service, accessTokenTTL: accessTokenTTL, cookieConfig: cookieConfig}
}

type registerRequest struct {
	Login    string `json:"login"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type registerResponse struct {
	ID    string `json:"id"`
	Login string `json:"login"`
	Email string `json:"email"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.service.Register(r.Context(), req.Login, req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrInvalidEmail), errors.Is(err, auth.ErrInvalidPassword), errors.Is(err, auth.ErrInvalidLogin):
			writeError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, auth.ErrEmailAlreadyExists):
			writeError(w, http.StatusConflict, err.Error())
		case errors.Is(err, auth.ErrLoginAlreadyExists):
			writeError(w, http.StatusConflict, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "failed to create user")
		}
		return
	}

	if !h.ensureCSRFCookie(w, r) {
		return
	}
	resp := registerResponse{ID: user.ID.String(), Login: user.Login, Email: user.Email}
	writeJSON(w, http.StatusCreated, resp)
}

type loginRequest struct {
	Identifier string `json:"identifier"`
	Email      string `json:"email"` // backwards compatibility
	Login      string `json:"login"`
	Password   string `json:"password"`
}

type loginResponse struct {
	ID    string `json:"id"`
	Login string `json:"login"`
	Email string `json:"email"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	identifier := strings.TrimSpace(req.Identifier)
	if identifier == "" {
		identifier = strings.TrimSpace(req.Email)
	}
	if identifier == "" {
		identifier = strings.TrimSpace(req.Login)
	}
	if identifier == "" {
		writeError(w, http.StatusBadRequest, "identifier is required")
		return
	}

	user, token, err := h.service.Login(r.Context(), identifier, req.Password)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to login")
		return
	}

	h.setAccessCookie(w, token)
	if !h.ensureCSRFCookie(w, r) {
		return
	}
	resp := loginResponse{ID: user.ID.String(), Login: user.Login, Email: user.Email}
	writeJSON(w, http.StatusOK, resp)
}

type meResponse struct {
	ID    string `json:"id"`
	Login string `json:"login"`
	Email string `json:"email"`
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if !h.ensureCSRFCookie(w, r) {
		return
	}
	resp := meResponse{ID: user.ID.String(), Login: user.Login, Email: user.Email}
	writeJSON(w, http.StatusOK, resp)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	h.clearAccessCookie(w)
	h.clearCSRFCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) setAccessCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, h.buildCookie(http.Cookie{
		Name:     auth.AccessCookieName,
		Value:    token,
		HttpOnly: true,
		MaxAge:   int(h.accessTokenTTL.Seconds()),
	}))
}

func (h *AuthHandler) clearAccessCookie(w http.ResponseWriter) {
	http.SetCookie(w, h.buildCookie(http.Cookie{
		Name:     auth.AccessCookieName,
		Value:    "",
		HttpOnly: true,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	}))
}

func (h *AuthHandler) ensureCSRFCookie(w http.ResponseWriter, r *http.Request) bool {
	if existing, err := r.Cookie(auth.CSRFCookieName); err == nil && existing.Value != "" {
		return true
	}

	token, err := generateCSRFToken()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate csrf token")
		return false
	}

	http.SetCookie(w, h.buildCookie(http.Cookie{
		Name:     auth.CSRFCookieName,
		Value:    token,
		HttpOnly: false,
		MaxAge:   int(h.accessTokenTTL.Seconds()),
	}))
	return true
}

func (h *AuthHandler) clearCSRFCookie(w http.ResponseWriter) {
	http.SetCookie(w, h.buildCookie(http.Cookie{
		Name:     auth.CSRFCookieName,
		Value:    "",
		HttpOnly: false,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	}))
}

func (h *AuthHandler) buildCookie(base http.Cookie) *http.Cookie {
	cookie := base
	cookie.Path = h.cookieConfig.Path
	cookie.Domain = h.cookieConfig.Domain
	cookie.SameSite = h.cookieConfig.SameSite
	cookie.Secure = h.cookieConfig.Secure
	return &cookie
}

func generateCSRFToken() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

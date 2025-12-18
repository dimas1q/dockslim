package auth

import (
	"errors"
	"net/http"
	"strings"
)

type Middleware struct {
	tokens *TokenManager
	users  UserStore
}

func NewMiddleware(tokens *TokenManager, users UserStore) *Middleware {
	return &Middleware{tokens: tokens, users: users}
}

func (m *Middleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString, err := extractBearerToken(r.Header.Get("Authorization"))
		if err != nil {
			respondJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		claims, err := m.tokens.ValidateToken(r.Context(), tokenString)
		if err != nil {
			respondJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		userID := claims.Subject
		if userID == "" {
			respondJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		user, err := m.users.GetUserByID(r.Context(), userID)
		if err != nil {
			respondJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		ctx := WithUser(r.Context(), user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func extractBearerToken(header string) (string, error) {
	if header == "" {
		return "", errors.New("missing authorization header")
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", errors.New("invalid authorization header")
	}

	if parts[1] == "" {
		return "", errors.New("empty bearer token")
	}

	return parts[1], nil
}

func respondJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(`{"error":"` + message + `"}`))
}

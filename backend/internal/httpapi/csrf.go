package httpapi

import (
	"net/http"
	"strings"

	"github.com/dimas1q/dockslim/backend/internal/auth"
)

const csrfHeaderName = "X-CSRF-Token"

func csrfMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions || r.URL.Path == "/health" {
			next.ServeHTTP(w, r)
			return
		}

		if r.Method != http.MethodPost &&
			r.Method != http.MethodPut &&
			r.Method != http.MethodPatch &&
			r.Method != http.MethodDelete {
			next.ServeHTTP(w, r)
			return
		}

		if r.URL.Path == "/api/v1/auth/login" || r.URL.Path == "/api/v1/auth/register" {
			next.ServeHTTP(w, r)
			return
		}

		if hasBearerAuth(r.Header.Get("Authorization")) {
			next.ServeHTTP(w, r)
			return
		}

		if cookie, err := r.Cookie(auth.AccessCookieName); err != nil || cookie.Value == "" {
			next.ServeHTTP(w, r)
			return
		}

		csrfCookie, err := r.Cookie(auth.CSRFCookieName)
		if err != nil || csrfCookie.Value == "" {
			writeError(w, http.StatusForbidden, "csrf validation failed")
			return
		}

		headerToken := r.Header.Get(csrfHeaderName)
		if headerToken == "" || headerToken != csrfCookie.Value {
			writeError(w, http.StatusForbidden, "csrf validation failed")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func hasBearerAuth(header string) bool {
	if header == "" {
		return false
	}
	parts := strings.SplitN(header, " ", 2)
	return len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") && parts[1] != ""
}

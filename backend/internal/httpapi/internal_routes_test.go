package httpapi

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dimas1q/dockslim/backend/internal/auth"
)

func TestNonAdminCannotAccessInternalRoutes(t *testing.T) {
	userStore := newMemoryUserStore()
	tokenStore := newMemoryKeyStore()
	tokenManager := newTokenManager(t, tokenStore)
	authService := auth.NewService(userStore, tokenManager)
	authHandler := NewAuthHandler(authService, time.Hour, CookieConfig{
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})
	authMiddleware := auth.NewMiddleware(tokenManager, userStore, nil, nil)
	accountHandler := NewAccountHandler(authService, &apiTokenServiceStub{}, AccountHandlerOptions{
		SubscriptionService:       &subscriptionServiceStub{},
		InternalSubscriptionToken: "internal-secret",
	})

	router := NewRouter(Dependencies{
		AuthHandler:    authHandler,
		AuthMiddleware: authMiddleware,
		AccountHandler: accountHandler,
	})

	createUser(t, userStore, "user1", "user1@example.com", "password123")

	loginRecorder := httptest.NewRecorder()
	loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(`{"email":"user1@example.com","password":"password123"}`))
	router.ServeHTTP(loginRecorder, loginReq)

	if loginRecorder.Code != http.StatusOK {
		t.Fatalf("expected login 200, got %d", loginRecorder.Code)
	}

	accessCookie := findCookie(loginRecorder.Result().Cookies(), auth.AccessCookieName)
	csrfCookie := findCookie(loginRecorder.Result().Cookies(), auth.CSRFCookieName)
	if accessCookie == nil || csrfCookie == nil {
		t.Fatalf("expected auth cookies")
	}

	body := bytes.NewBufferString(`{"user_id":"00000000-0000-0000-0000-000000000001","plan_id":"pro","status":"active"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/internal/subscriptions", body)
	req.AddCookie(accessCookie)
	req.AddCookie(csrfCookie)
	req.Header.Set(csrfHeaderName, csrfCookie.Value)
	req.Header.Set("X-DockSlim-Internal-Token", "internal-secret")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

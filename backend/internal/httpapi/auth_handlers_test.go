package httpapi

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dimas1q/dockslim/backend/internal/auth"
)

func TestAuthHandlerRegisterConflictJSON(t *testing.T) {
	service := auth.NewService(&conflictUserStore{}, &noopTokenIssuer{})
	handler := NewAuthHandler(service, time.Hour, CookieConfig{
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString(`{"login":"user","email":"user@example.com","password":"password123"}`))

	handler.Register(recorder, req)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("expected status 409, got %d", recorder.Code)
	}

	if contentType := recorder.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("expected JSON content type, got %s", contentType)
	}

	expected := `{"error":"email already registered"}
`
	if body := recorder.Body.String(); body != expected {
		t.Fatalf("expected body %q, got %q", expected, body)
	}
}

type conflictUserStore struct{}

func (c *conflictUserStore) CreateUser(ctx context.Context, login, email, passwordHash string) (auth.User, error) {
	return auth.User{}, auth.ErrEmailAlreadyExists
}

func (c *conflictUserStore) GetUserByEmail(ctx context.Context, email string) (auth.User, error) {
	return auth.User{}, auth.ErrUserNotFound
}

func (c *conflictUserStore) GetUserByLogin(ctx context.Context, login string) (auth.User, error) {
	return auth.User{}, auth.ErrUserNotFound
}

func (c *conflictUserStore) GetUserByID(ctx context.Context, id string) (auth.User, error) {
	return auth.User{}, auth.ErrUserNotFound
}

func (c *conflictUserStore) UpdateUserProfile(ctx context.Context, id, login, email string) (auth.User, error) {
	return auth.User{}, auth.ErrEmailAlreadyExists
}

type noopTokenIssuer struct{}

func (n *noopTokenIssuer) GenerateAccessToken(ctx context.Context, user auth.User) (string, error) {
	return "", nil
}

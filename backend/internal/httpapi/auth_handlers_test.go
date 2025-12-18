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
	handler := NewAuthHandler(service, time.Hour)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString(`{"email":"user@example.com","password":"password123"}`))

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

func (c *conflictUserStore) CreateUser(ctx context.Context, email, passwordHash string) (auth.User, error) {
	return auth.User{}, auth.ErrEmailAlreadyExists
}

func (c *conflictUserStore) GetUserByEmail(ctx context.Context, email string) (auth.User, error) {
	return auth.User{}, auth.ErrUserNotFound
}

func (c *conflictUserStore) GetUserByID(ctx context.Context, id string) (auth.User, error) {
	return auth.User{}, auth.ErrUserNotFound
}

type noopTokenIssuer struct{}

func (n *noopTokenIssuer) GenerateAccessToken(ctx context.Context, user auth.User) (string, error) {
	return "", nil
}

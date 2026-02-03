package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dimas1q/dockslim/backend/internal/apitokens"
	"github.com/dimas1q/dockslim/backend/internal/auth"
	"github.com/google/uuid"
)

type apiTokenServiceStub struct {
	tokens []apitokens.Token
	err    error
}

func (s *apiTokenServiceStub) CreateToken(ctx context.Context, userID uuid.UUID, name string, expiresAt *time.Time) (apitokens.Token, string, error) {
	if s.err != nil {
		return apitokens.Token{}, "", s.err
	}
	token := apitokens.Token{ID: uuid.New(), UserID: userID, Name: name, CreatedAt: time.Now()}
	s.tokens = append(s.tokens, token)
	return token, "ds_api_token", nil
}

func (s *apiTokenServiceStub) ListTokens(ctx context.Context, userID uuid.UUID) ([]apitokens.Token, error) {
	return s.tokens, s.err
}

func (s *apiTokenServiceStub) RevokeToken(ctx context.Context, userID, tokenID uuid.UUID) error {
	return s.err
}

func TestCreateAPITokenConflictReturns409(t *testing.T) {
	user := auth.User{ID: uuid.New()}
	service := &apiTokenServiceStub{err: apitokens.ErrNameConflict}
	handler := NewAccountHandler(nil, service)

	body, _ := json.Marshal(map[string]string{"name": "personal"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/account/api-tokens", bytes.NewBuffer(body))
	req = req.WithContext(auth.WithUser(req.Context(), user))
	rec := httptest.NewRecorder()

	handler.CreateAPIToken(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rec.Code)
	}
}

func TestRevokeAPITokenNotFound(t *testing.T) {
	user := auth.User{ID: uuid.New()}
	service := &apiTokenServiceStub{err: apitokens.ErrTokenNotFound}
	handler := NewAccountHandler(nil, service)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/account/api-tokens/token/revoke", nil)
	req = req.WithContext(auth.WithUser(req.Context(), user))
	tokenID := uuid.New().String()
	req = withURLParam(req, "tokenId", tokenID)
	rec := httptest.NewRecorder()

	handler.RevokeAPIToken(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestUpdateProfileInvalidEmail(t *testing.T) {
	user := auth.User{ID: uuid.New(), Login: "demo", Email: "demo@example.com"}
	userStore := newMemoryUserStore()
	userStore.usersByID[user.ID.String()] = user
	userStore.usersByEmail[user.Email] = user
	userStore.usersByLogin[user.Login] = user
	tokenStore := newMemoryKeyStore()
	tokenManager := newTokenManager(t, tokenStore)
	authService := auth.NewService(userStore, tokenManager)
	handler := NewAccountHandler(authService, &apiTokenServiceStub{})

	body, _ := json.Marshal(map[string]string{"email": "bad-email"})
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/account/me", bytes.NewBuffer(body))
	req = req.WithContext(auth.WithUser(req.Context(), user))
	rec := httptest.NewRecorder()

	handler.UpdateProfile(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

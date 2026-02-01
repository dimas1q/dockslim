package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dimas1q/dockslim/backend/internal/auth"
	"github.com/dimas1q/dockslim/backend/internal/citokens"
	"github.com/google/uuid"
)

type ciTokenServiceStub struct {
	tokens []citokens.Token
	err    error
}

func (s *ciTokenServiceStub) CreateToken(ctx context.Context, userID, projectID uuid.UUID, name string, expiresAt *time.Time) (citokens.Token, string, error) {
	if s.err != nil {
		return citokens.Token{}, "", s.err
	}
	token := citokens.Token{ID: uuid.New(), ProjectID: projectID, Name: name, CreatedAt: time.Now()}
	s.tokens = append(s.tokens, token)
	return token, "ds_ci_token", nil
}

func (s *ciTokenServiceStub) ListTokens(ctx context.Context, userID, projectID uuid.UUID) ([]citokens.Token, error) {
	return s.tokens, s.err
}

func (s *ciTokenServiceStub) RevokeToken(ctx context.Context, userID, projectID, tokenID uuid.UUID) error {
	return s.err
}

func TestCITokenCreateConflictReturns409(t *testing.T) {
	projectID := uuid.New()
	user := auth.User{ID: uuid.New()}
	service := &ciTokenServiceStub{err: citokens.ErrNameConflict}
	handler := NewCITokensHandler(service)

	payload := map[string]string{"name": "CI"}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/"+projectID.String()+"/ci-tokens", bytes.NewBuffer(body))
	req = req.WithContext(auth.WithUser(req.Context(), user))
	req = withURLParam(req, "projectId", projectID.String())
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rec.Code)
	}
}

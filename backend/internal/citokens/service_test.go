package citokens

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/dimas1q/dockslim/backend/internal/projects"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type memRepo struct {
	tokens map[uuid.UUID]Token
}

func newMemRepo() *memRepo {
	return &memRepo{tokens: make(map[uuid.UUID]Token)}
}

func (m *memRepo) CreateToken(ctx context.Context, params CreateTokenParams) (Token, error) {
	token := Token{
		ID:        params.ID,
		ProjectID: params.ProjectID,
		Name:      params.Name,
		TokenHash: params.TokenHash,
		ExpiresAt: params.ExpiresAt,
		CreatedAt: time.Now(),
	}
	m.tokens[token.ID] = token
	return token, nil
}

func (m *memRepo) ListTokensByProject(ctx context.Context, projectID uuid.UUID) ([]Token, error) {
	var tokens []Token
	for _, t := range m.tokens {
		if t.ProjectID == projectID {
			tokens = append(tokens, t)
		}
	}
	return tokens, nil
}

func (m *memRepo) GetTokenByID(ctx context.Context, tokenID uuid.UUID) (Token, error) {
	token, ok := m.tokens[tokenID]
	if !ok {
		return Token{}, ErrTokenNotFound
	}
	return token, nil
}

func (m *memRepo) RevokeToken(ctx context.Context, projectID, tokenID uuid.UUID) error {
	token, ok := m.tokens[tokenID]
	if !ok {
		return ErrTokenNotFound
	}
	if token.ProjectID != projectID {
		return ErrTokenNotFound
	}
	now := time.Now()
	token.RevokedAt = &now
	m.tokens[tokenID] = token
	return nil
}

func (m *memRepo) UpdateLastUsed(ctx context.Context, tokenID uuid.UUID, ts time.Time) error {
	token, ok := m.tokens[tokenID]
	if !ok {
		return ErrTokenNotFound
	}
	token.LastUsedAt = &ts
	m.tokens[tokenID] = token
	return nil
}

type memberStub struct {
	role string
	err  error
}

func (m *memberStub) GetMemberRole(ctx context.Context, projectID, userID uuid.UUID) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.role, nil
}

func TestCreateAndAuthenticateToken(t *testing.T) {
	repo := newMemRepo()
	members := &memberStub{role: projects.RoleOwner}
	service := NewService(repo, members)
	now := time.Now()
	service.nowFn = func() time.Time { return now }

	projectID := uuid.New()
	userID := uuid.New()

	token, plain, err := service.CreateToken(context.Background(), userID, projectID, " CI Runner ", nil)
	if err != nil {
		t.Fatalf("CreateToken returned error: %v", err)
	}

	if plain == "" || !strings.HasPrefix(plain, TokenPrefix) {
		t.Fatalf("expected plaintext token with prefix, got %s", plain)
	}

	stored, ok := repo.tokens[token.ID]
	if !ok {
		t.Fatalf("token not persisted in repo")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(stored.TokenHash), []byte(strings.SplitN(plain[len(TokenPrefix):], "_", 2)[1])); err != nil {
		t.Fatalf("stored hash does not match token secret: %v", err)
	}

	if stored.Name != "CI Runner" {
		t.Fatalf("expected cleaned name, got %s", stored.Name)
	}

	authToken, err := service.Authenticate(context.Background(), plain)
	if err != nil {
		t.Fatalf("Authenticate returned error: %v", err)
	}
	if authToken.ID != token.ID || authToken.ProjectID != projectID {
		t.Fatalf("authenticated token mismatch")
	}
	if authToken.LastUsedAt == nil || !authToken.LastUsedAt.Equal(now) {
		t.Fatalf("expected last_used_at to be updated")
	}
}

func TestAuthenticateRevokedTokenRejected(t *testing.T) {
	repo := newMemRepo()
	members := &memberStub{role: projects.RoleOwner}
	service := NewService(repo, members)

	projectID := uuid.New()
	userID := uuid.New()

	token, plain, err := service.CreateToken(context.Background(), userID, projectID, "deploy", nil)
	if err != nil {
		t.Fatalf("CreateToken returned error: %v", err)
	}

	now := time.Now()
	token.RevokedAt = &now
	repo.tokens[token.ID] = token

	if _, err := service.Authenticate(context.Background(), plain); !errors.Is(err, ErrRevokedToken) {
		t.Fatalf("expected revoked token error, got %v", err)
	}
}

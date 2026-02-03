package apitokens

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

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
		UserID:    params.UserID,
		Name:      params.Name,
		TokenHash: params.TokenHash,
		ExpiresAt: params.ExpiresAt,
		CreatedAt: time.Now(),
	}
	m.tokens[token.ID] = token
	return token, nil
}

func (m *memRepo) ListTokensByUser(ctx context.Context, userID uuid.UUID) ([]Token, error) {
	var tokens []Token
	for _, t := range m.tokens {
		if t.UserID == userID {
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

func (m *memRepo) RevokeToken(ctx context.Context, userID, tokenID uuid.UUID) error {
	token, ok := m.tokens[tokenID]
	if !ok {
		return ErrTokenNotFound
	}
	if token.UserID != userID {
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

func TestCreateAndAuthenticateToken(t *testing.T) {
	repo := newMemRepo()
	service := NewService(repo)
	now := time.Now()
	service.nowFn = func() time.Time { return now }

	userID := uuid.New()

	token, plain, err := service.CreateToken(context.Background(), userID, " Personal Token ", nil)
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
	parts := strings.SplitN(plain[len(TokenPrefix):], "_", 2)
	if len(parts) != 2 {
		t.Fatalf("unexpected token format")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(stored.TokenHash), []byte(parts[1])); err != nil {
		t.Fatalf("stored hash does not match token secret: %v", err)
	}

	if stored.Name != "Personal Token" {
		t.Fatalf("expected cleaned name, got %s", stored.Name)
	}

	authToken, err := service.Authenticate(context.Background(), plain)
	if err != nil {
		t.Fatalf("Authenticate returned error: %v", err)
	}
	if authToken.ID != token.ID || authToken.UserID != userID {
		t.Fatalf("authenticated token mismatch")
	}
	if authToken.LastUsedAt == nil || !authToken.LastUsedAt.Equal(now) {
		t.Fatalf("expected last_used_at to be updated")
	}
}

func TestAuthenticateRevokedTokenRejected(t *testing.T) {
	repo := newMemRepo()
	service := NewService(repo)

	userID := uuid.New()

	token, plain, err := service.CreateToken(context.Background(), userID, "deploy", nil)
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

func TestRevokeTokenEnforcesOwnership(t *testing.T) {
	repo := newMemRepo()
	service := NewService(repo)

	userID := uuid.New()
	otherUser := uuid.New()

	token, _, err := service.CreateToken(context.Background(), userID, "deploy", nil)
	if err != nil {
		t.Fatalf("CreateToken returned error: %v", err)
	}

	if err := service.RevokeToken(context.Background(), otherUser, token.ID); !errors.Is(err, ErrTokenNotFound) {
		t.Fatalf("expected ownership enforcement error, got %v", err)
	}
}

package apitokens

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidName  = errors.New("invalid api token name")
	ErrInvalidToken = errors.New("invalid api token")
	ErrRevokedToken = errors.New("api token revoked")
	ErrExpiredToken = errors.New("api token expired")
	ErrNameConflict = errors.New("api token name already exists")
)

type RepositoryStore interface {
	CreateToken(ctx context.Context, params CreateTokenParams) (Token, error)
	ListTokensByUser(ctx context.Context, userID uuid.UUID) ([]Token, error)
	GetTokenByID(ctx context.Context, tokenID uuid.UUID) (Token, error)
	RevokeToken(ctx context.Context, userID, tokenID uuid.UUID) error
	UpdateLastUsed(ctx context.Context, tokenID uuid.UUID, ts time.Time) error
}

type Service struct {
	repo  RepositoryStore
	nowFn func() time.Time
}

func NewService(repo RepositoryStore) *Service {
	return &Service{
		repo:  repo,
		nowFn: time.Now,
	}
}

func (s *Service) CreateToken(ctx context.Context, userID uuid.UUID, name string, expiresAt *time.Time) (Token, string, error) {
	cleanName := strings.TrimSpace(name)
	if len(cleanName) == 0 || len(cleanName) > 100 {
		return Token{}, "", ErrInvalidName
	}

	tokenID := uuid.New()
	secret, err := generateSecret()
	if err != nil {
		return Token{}, "", err
	}

	hashBytes, err := bcrypt.GenerateFromPassword([]byte(secret), bcrypt.DefaultCost)
	if err != nil {
		return Token{}, "", fmt.Errorf("failed to hash token: %w", err)
	}

	plain := fmt.Sprintf("%s%s_%s", TokenPrefix, tokenID.String(), secret)

	params := CreateTokenParams{
		ID:        tokenID,
		UserID:    userID,
		Name:      cleanName,
		TokenHash: string(hashBytes),
		ExpiresAt: expiresAt,
	}

	token, err := s.repo.CreateToken(ctx, params)
	if err != nil {
		if errors.Is(err, ErrTokenConflict) {
			return Token{}, "", ErrNameConflict
		}
		return Token{}, "", err
	}

	return token, plain, nil
}

func (s *Service) ListTokens(ctx context.Context, userID uuid.UUID) ([]Token, error) {
	return s.repo.ListTokensByUser(ctx, userID)
}

func (s *Service) RevokeToken(ctx context.Context, userID, tokenID uuid.UUID) error {
	return s.repo.RevokeToken(ctx, userID, tokenID)
}

// Authenticate validates a user API token string and returns its metadata.
func (s *Service) Authenticate(ctx context.Context, tokenString string) (Token, error) {
	tokenID, secret, err := parsePlainToken(tokenString)
	if err != nil {
		return Token{}, ErrInvalidToken
	}

	token, err := s.repo.GetTokenByID(ctx, tokenID)
	if err != nil {
		if errors.Is(err, ErrTokenNotFound) {
			return Token{}, ErrInvalidToken
		}
		return Token{}, err
	}

	if token.RevokedAt != nil {
		return Token{}, ErrRevokedToken
	}
	if token.ExpiresAt != nil && s.nowFn().After(*token.ExpiresAt) {
		return Token{}, ErrExpiredToken
	}

	if err := bcrypt.CompareHashAndPassword([]byte(token.TokenHash), []byte(secret)); err != nil {
		return Token{}, ErrInvalidToken
	}

	now := s.nowFn()
	_ = s.repo.UpdateLastUsed(ctx, token.ID, now)
	token.LastUsedAt = &now

	return token, nil
}

func generateSecret() (string, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("failed to read random bytes: %w", err)
	}
	return hex.EncodeToString(buf), nil
}

func parsePlainToken(raw string) (uuid.UUID, string, error) {
	if !strings.HasPrefix(raw, TokenPrefix) {
		return uuid.Nil, "", ErrInvalidToken
	}
	withoutPrefix := strings.TrimPrefix(raw, TokenPrefix)
	parts := strings.SplitN(withoutPrefix, "_", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return uuid.Nil, "", ErrInvalidToken
	}
	id, err := uuid.Parse(parts[0])
	if err != nil {
		return uuid.Nil, "", ErrInvalidToken
	}
	return id, parts[1], nil
}

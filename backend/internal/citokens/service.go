package citokens

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/dimas1q/dockslim/backend/internal/projects"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidName     = errors.New("invalid token name")
	ErrProjectNotFound = errors.New("project not found")
	ErrNotOwner        = errors.New("user is not project owner")
	ErrInvalidToken    = errors.New("invalid ci token")
	ErrRevokedToken    = errors.New("ci token revoked")
	ErrExpiredToken    = errors.New("ci token expired")
	ErrNameConflict    = errors.New("ci token name already exists")
)

type RepositoryStore interface {
	CreateToken(ctx context.Context, params CreateTokenParams) (Token, error)
	ListTokensByProject(ctx context.Context, projectID uuid.UUID) ([]Token, error)
	GetTokenByID(ctx context.Context, tokenID uuid.UUID) (Token, error)
	RevokeToken(ctx context.Context, projectID, tokenID uuid.UUID) error
	UpdateLastUsed(ctx context.Context, tokenID uuid.UUID, ts time.Time) error
}

type MembershipStore interface {
	GetMemberRole(ctx context.Context, projectID, userID uuid.UUID) (string, error)
}

type Service struct {
	repo    RepositoryStore
	members MembershipStore
	nowFn   func() time.Time
}

func NewService(repo RepositoryStore, members MembershipStore) *Service {
	return &Service{
		repo:    repo,
		members: members,
		nowFn:   time.Now,
	}
}

func (s *Service) CreateToken(ctx context.Context, userID, projectID uuid.UUID, name string, expiresAt *time.Time) (Token, string, error) {
	if err := s.ensureOwner(ctx, userID, projectID); err != nil {
		return Token{}, "", err
	}

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
		ProjectID: projectID,
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

func (s *Service) ListTokens(ctx context.Context, userID, projectID uuid.UUID) ([]Token, error) {
	if err := s.ensureOwner(ctx, userID, projectID); err != nil {
		return nil, err
	}
	return s.repo.ListTokensByProject(ctx, projectID)
}

func (s *Service) RevokeToken(ctx context.Context, userID, projectID, tokenID uuid.UUID) error {
	if err := s.ensureOwner(ctx, userID, projectID); err != nil {
		return err
	}
	return s.repo.RevokeToken(ctx, projectID, tokenID)
}

// Authenticate validates a CI token string and returns its metadata.
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

func (s *Service) ensureOwner(ctx context.Context, userID, projectID uuid.UUID) error {
	if s.members == nil {
		return ErrNotOwner
	}
	role, err := s.members.GetMemberRole(ctx, projectID, userID)
	if err != nil {
		if errors.Is(err, projects.ErrProjectMemberNotFound) {
			return ErrProjectNotFound
		}
		return err
	}
	if role != projects.RoleOwner {
		return ErrNotOwner
	}
	return nil
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

package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrMissingKID    = errors.New("token missing kid header")
	ErrUnknownKey    = errors.New("unknown signing key")
	ErrInvalidToken  = errors.New("invalid token")
	ErrUnexpectedAlg = errors.New("unexpected signing algorithm")
)

const (
	DefaultAccessTokenTTL = 24 * time.Hour
	DefaultSigningAlg     = "HS256"
)

type KeyStore interface {
	ListActiveKeys(ctx context.Context) ([]AuthKey, error)
	GetKeyByID(ctx context.Context, keyID string) (AuthKey, error)
	CreateAuthKey(ctx context.Context, key AuthKey) (AuthKey, error)
}

type AccessTokenClaims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

type TokenManager struct {
	store          KeyStore
	accessTokenTTL time.Duration

	mu           sync.RWMutex
	keys         map[string]AuthKey
	signingKeyID string
}

func NewTokenManager(ctx context.Context, store KeyStore, accessTokenTTL time.Duration) (*TokenManager, error) {
	manager := &TokenManager{
		store:          store,
		accessTokenTTL: accessTokenTTL,
		keys:           make(map[string]AuthKey),
	}

	if err := manager.ensureSigningKey(ctx); err != nil {
		return nil, err
	}

	return manager, nil
}

func (m *TokenManager) ensureSigningKey(ctx context.Context) error {
	keys, err := m.store.ListActiveKeys(ctx)
	if err != nil {
		return err
	}

	if len(keys) == 0 {
		signingKey, err := generateSigningKey()
		if err != nil {
			return err
		}

		newKey := AuthKey{
			KeyID:      uuid.NewString(),
			SigningKey: signingKey,
			Algorithm:  DefaultSigningAlg,
			IsActive:   true,
			CreatedAt:  time.Now().UTC(),
		}
		created, err := m.store.CreateAuthKey(ctx, newKey)
		if err != nil {
			return err
		}
		keys = append(keys, created)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for _, key := range keys {
		m.keys[key.KeyID] = key
	}

	m.signingKeyID = keys[0].KeyID
	return nil
}

func (m *TokenManager) GenerateAccessToken(ctx context.Context, user User) (string, error) {
	key, err := m.currentSigningKey(ctx)
	if err != nil {
		return "", err
	}

	signingMethod := jwt.GetSigningMethod(key.Algorithm)
	if signingMethod == nil {
		return "", ErrUnexpectedAlg
	}

	now := time.Now().UTC()
	claims := AccessTokenClaims{
		Email: user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.accessTokenTTL)),
		},
	}

	token := jwt.NewWithClaims(signingMethod, claims)
	token.Header["kid"] = key.KeyID

	return token.SignedString([]byte(key.SigningKey))
}

func (m *TokenManager) ValidateToken(ctx context.Context, tokenString string) (*AccessTokenClaims, error) {
	claims := &AccessTokenClaims{}

	parsed, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		kid, ok := token.Header["kid"].(string)
		if !ok || kid == "" {
			return nil, ErrMissingKID
		}

		key, err := m.findKey(ctx, kid)
		if err != nil {
			return nil, err
		}

		if token.Method.Alg() != key.Algorithm {
			return nil, ErrUnexpectedAlg
		}

		return []byte(key.SigningKey), nil
	})
	if err != nil {
		return nil, err
	}

	if !parsed.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func (m *TokenManager) findKey(ctx context.Context, keyID string) (AuthKey, error) {
	m.mu.RLock()
	key, ok := m.keys[keyID]
	m.mu.RUnlock()

	if ok {
		return key, nil
	}

	fetched, err := m.store.GetKeyByID(ctx, keyID)
	if err != nil {
		if errors.Is(err, ErrAuthKeyNotFound) {
			return AuthKey{}, ErrUnknownKey
		}
		return AuthKey{}, err
	}

	m.mu.Lock()
	m.keys[keyID] = fetched
	m.mu.Unlock()

	return fetched, nil
}

func (m *TokenManager) currentSigningKey(ctx context.Context) (AuthKey, error) {
	m.mu.RLock()
	keyID := m.signingKeyID
	key, ok := m.keys[keyID]
	m.mu.RUnlock()

	if ok {
		return key, nil
	}

	if err := m.ensureSigningKey(ctx); err != nil {
		return AuthKey{}, err
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	key, ok = m.keys[m.signingKeyID]
	if !ok {
		return AuthKey{}, ErrUnknownKey
	}

	return key, nil
}

func generateSigningKey() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("failed to read random bytes: %w", err)
	}
	return base64.StdEncoding.EncodeToString(buf), nil
}

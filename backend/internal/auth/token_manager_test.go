package auth

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestTokenManagerGeneratesKeyWhenMissing(t *testing.T) {
	store := newMemoryKeyStore()

	_, err := NewTokenManager(context.Background(), store, time.Hour)
	if err != nil {
		t.Fatalf("expected token manager creation to succeed, got %v", err)
	}

	if len(store.keys) == 0 {
		t.Fatalf("expected a signing key to be created")
	}
}

func TestTokenManagerSignsAndValidatesToken(t *testing.T) {
	store := newMemoryKeyStore()
	manager, err := NewTokenManager(context.Background(), store, time.Hour)
	if err != nil {
		t.Fatalf("expected token manager creation to succeed, got %v", err)
	}

	user := User{ID: uuid.New(), Email: "user@example.com"}
	tokenString, err := manager.GenerateAccessToken(context.Background(), user)
	if err != nil {
		t.Fatalf("expected token generation to succeed, got %v", err)
	}

	parser := jwt.NewParser()
	parsed, _, err := parser.ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		t.Fatalf("failed to parse token header: %v", err)
	}

	kid, ok := parsed.Header["kid"].(string)
	if !ok || kid == "" {
		t.Fatalf("expected kid header to be set")
	}

	claims, err := manager.ValidateToken(context.Background(), tokenString)
	if err != nil {
		t.Fatalf("expected token validation to succeed, got %v", err)
	}

	if claims.Subject != user.ID.String() {
		t.Fatalf("expected subject %s, got %s", user.ID.String(), claims.Subject)
	}

	if claims.Email != user.Email {
		t.Fatalf("expected email %s, got %s", user.Email, claims.Email)
	}

	if _, exists := store.keys[kid]; !exists {
		t.Fatalf("expected kid %s to exist in keystore", kid)
	}
}

type memoryKeyStore struct {
	keys   map[string]AuthKey
	active []AuthKey
}

func newMemoryKeyStore() *memoryKeyStore {
	return &memoryKeyStore{keys: make(map[string]AuthKey)}
}

func (m *memoryKeyStore) ListActiveKeys(ctx context.Context) ([]AuthKey, error) {
	return append([]AuthKey{}, m.active...), nil
}

func (m *memoryKeyStore) GetKeyByID(ctx context.Context, keyID string) (AuthKey, error) {
	key, ok := m.keys[keyID]
	if !ok {
		return AuthKey{}, ErrAuthKeyNotFound
	}
	return key, nil
}

func (m *memoryKeyStore) CreateAuthKey(ctx context.Context, key AuthKey) (AuthKey, error) {
	if key.ID == uuid.Nil {
		key.ID = uuid.New()
	}
	m.keys[key.KeyID] = key
	m.active = append([]AuthKey{key}, m.active...)
	return key, nil
}

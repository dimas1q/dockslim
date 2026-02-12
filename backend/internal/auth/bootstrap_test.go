package auth

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

type bootstrapUserStoreStub struct {
	hasAdmin bool

	byEmail map[string]User
	byID    map[string]User

	createCalls int
	setCalls    int
}

func (s *bootstrapUserStoreStub) HasAnyAdmin(ctx context.Context) (bool, error) {
	return s.hasAdmin, nil
}

func (s *bootstrapUserStoreStub) GetUserByEmail(ctx context.Context, email string) (User, error) {
	if user, ok := s.byEmail[email]; ok {
		return user, nil
	}
	return User{}, ErrUserNotFound
}

func (s *bootstrapUserStoreStub) CreateUser(ctx context.Context, login, email, passwordHash string) (User, error) {
	s.createCalls++
	if _, exists := s.byEmail[email]; exists {
		return User{}, ErrEmailAlreadyExists
	}
	user := User{
		ID:           uuid.New(),
		Login:        login,
		Email:        email,
		PasswordHash: passwordHash,
		IsAdmin:      false,
	}
	s.byEmail[email] = user
	s.byID[user.ID.String()] = user
	return user, nil
}

func (s *bootstrapUserStoreStub) SetUserAdmin(ctx context.Context, id string, isAdmin bool) error {
	s.setCalls++
	user, ok := s.byID[id]
	if !ok {
		return ErrUserNotFound
	}
	user.IsAdmin = isAdmin
	s.byID[id] = user
	s.byEmail[user.Email] = user
	return nil
}

func TestEnsureBootstrapAdminSkipsWhenAdminExists(t *testing.T) {
	store := &bootstrapUserStoreStub{
		hasAdmin: true,
		byEmail:  map[string]User{},
		byID:     map[string]User{},
	}

	changed, err := EnsureBootstrapAdmin(context.Background(), store, BootstrapAdminConfig{
		Email:    "admin@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if changed {
		t.Fatalf("expected no changes when admin already exists")
	}
	if store.createCalls != 0 {
		t.Fatalf("expected no user creation calls, got %d", store.createCalls)
	}
}

func TestEnsureBootstrapAdminDoesNothingWithoutExplicitConfig(t *testing.T) {
	store := &bootstrapUserStoreStub{
		hasAdmin: false,
		byEmail:  map[string]User{},
		byID:     map[string]User{},
	}

	changed, err := EnsureBootstrapAdmin(context.Background(), store, BootstrapAdminConfig{})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if changed {
		t.Fatalf("expected no admin bootstrap without explicit env config")
	}
	if store.createCalls != 0 {
		t.Fatalf("expected zero user create calls, got %d", store.createCalls)
	}
	if store.setCalls != 0 {
		t.Fatalf("expected zero user admin updates, got %d", store.setCalls)
	}
}

func TestEnsureBootstrapAdminCreatesAdminWhenMissing(t *testing.T) {
	store := &bootstrapUserStoreStub{
		hasAdmin: false,
		byEmail:  map[string]User{},
		byID:     map[string]User{},
	}

	changed, err := EnsureBootstrapAdmin(context.Background(), store, BootstrapAdminConfig{
		Email:    "admin@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !changed {
		t.Fatalf("expected bootstrap to create admin")
	}

	user, ok := store.byEmail["admin@example.com"]
	if !ok {
		t.Fatalf("expected bootstrap user to be created")
	}
	if !user.IsAdmin {
		t.Fatalf("expected created user to be admin")
	}
}

func TestEnsureBootstrapAdminUsesConfiguredUsername(t *testing.T) {
	store := &bootstrapUserStoreStub{
		hasAdmin: false,
		byEmail:  map[string]User{},
		byID:     map[string]User{},
	}

	changed, err := EnsureBootstrapAdmin(context.Background(), store, BootstrapAdminConfig{
		Email:    "admin@example.com",
		Username: "platform-admin",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !changed {
		t.Fatalf("expected bootstrap to create admin")
	}

	user, ok := store.byEmail["admin@example.com"]
	if !ok {
		t.Fatalf("expected bootstrap user to be created")
	}
	if user.Login != "platform-admin" {
		t.Fatalf("expected username platform-admin, got %q", user.Login)
	}
}

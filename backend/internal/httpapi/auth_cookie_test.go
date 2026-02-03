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
	"github.com/google/uuid"
)

func TestLoginSetsCookie(t *testing.T) {
	userStore := newMemoryUserStore()
	tokenStore := newMemoryKeyStore()
	tokenManager := newTokenManager(t, tokenStore)
	service := auth.NewService(userStore, tokenManager)
	handler := NewAuthHandler(service, time.Hour, CookieConfig{
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	user := createUser(t, userStore, "user", "user@example.com", "password123")

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(`{"email":"user@example.com","password":"password123"}`))

	handler.Login(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	cookie := recorder.Result().Cookies()
	if len(cookie) == 0 {
		t.Fatalf("expected Set-Cookie header to be present")
	}

	var accessCookie *http.Cookie
	for _, c := range cookie {
		if c.Name == auth.AccessCookieName {
			accessCookie = c
			break
		}
	}
	if accessCookie == nil {
		t.Fatalf("expected %s cookie to be set", auth.AccessCookieName)
	}
	if accessCookie.HttpOnly != true {
		t.Fatalf("expected cookie to be HttpOnly")
	}
	if accessCookie.SameSite != http.SameSiteLaxMode {
		t.Fatalf("expected SameSite=Lax")
	}

	var csrfCookie *http.Cookie
	for _, c := range cookie {
		if c.Name == auth.CSRFCookieName {
			csrfCookie = c
			break
		}
	}
	if csrfCookie == nil {
		t.Fatalf("expected %s cookie to be set", auth.CSRFCookieName)
	}
	if csrfCookie.HttpOnly {
		t.Fatalf("expected csrf cookie to not be HttpOnly")
	}

	var payload map[string]any
	if err := json.NewDecoder(recorder.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode response JSON: %v", err)
	}
	if _, ok := payload["access_token"]; ok {
		t.Fatalf("did not expect access_token in response body")
	}
	if payload["id"] != user.ID.String() {
		t.Fatalf("expected user id in response")
	}
}

func TestMeWithCookie(t *testing.T) {
	userStore := newMemoryUserStore()
	tokenStore := newMemoryKeyStore()
	tokenManager := newTokenManager(t, tokenStore)
	service := auth.NewService(userStore, tokenManager)
	middleware := auth.NewMiddleware(tokenManager, userStore, nil, nil)
	handler := NewAuthHandler(service, time.Hour, CookieConfig{
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	user := createUser(t, userStore, "me", "me@example.com", "password123")

	router := NewRouter(Dependencies{
		AuthHandler:     handler,
		AuthMiddleware:  middleware,
		ProjectsHandler: nil,
		AllowedOrigins:  nil,
	})

	loginRecorder := httptest.NewRecorder()
	loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(`{"email":"me@example.com","password":"password123"}`))
	router.ServeHTTP(loginRecorder, loginReq)

	if loginRecorder.Code != http.StatusOK {
		t.Fatalf("expected login status 200, got %d", loginRecorder.Code)
	}

	accessCookie := findAccessCookie(loginRecorder.Result().Cookies())
	if accessCookie == nil {
		t.Fatalf("expected access cookie to be set")
	}

	meRecorder := httptest.NewRecorder()
	meReq := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	meReq.AddCookie(accessCookie)
	router.ServeHTTP(meRecorder, meReq)

	if meRecorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", meRecorder.Code)
	}

	var resp map[string]string
	if err := json.NewDecoder(meRecorder.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["id"] != user.ID.String() {
		t.Fatalf("expected user id to match")
	}
}

func TestMeWithoutCookieUnauthorized(t *testing.T) {
	userStore := newMemoryUserStore()
	tokenStore := newMemoryKeyStore()
	tokenManager := newTokenManager(t, tokenStore)
	service := auth.NewService(userStore, tokenManager)
	middleware := auth.NewMiddleware(tokenManager, userStore, nil, nil)
	handler := NewAuthHandler(service, time.Hour, CookieConfig{
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	router := NewRouter(Dependencies{
		AuthHandler:     handler,
		AuthMiddleware:  middleware,
		ProjectsHandler: nil,
		AllowedOrigins:  nil,
	})

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", recorder.Code)
	}
	expectedBody := "{\"error\":\"unauthorized\"}\n"
	if recorder.Body.String() != expectedBody {
		t.Fatalf("expected body %s, got %s", expectedBody, recorder.Body.String())
	}
}

type memoryUserStore struct {
	usersByID    map[string]auth.User
	usersByEmail map[string]auth.User
	usersByLogin map[string]auth.User
}

func newMemoryUserStore() *memoryUserStore {
	return &memoryUserStore{
		usersByID:    make(map[string]auth.User),
		usersByEmail: make(map[string]auth.User),
		usersByLogin: make(map[string]auth.User),
	}
}

func (m *memoryUserStore) CreateUser(ctx context.Context, login, email, passwordHash string) (auth.User, error) {
	user := auth.User{
		ID:           uuid.New(),
		Login:        login,
		Email:        email,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now().UTC(),
	}
	m.usersByID[user.ID.String()] = user
	m.usersByEmail[email] = user
	m.usersByLogin[login] = user
	return user, nil
}

func (m *memoryUserStore) GetUserByEmail(ctx context.Context, email string) (auth.User, error) {
	user, ok := m.usersByEmail[email]
	if !ok {
		return auth.User{}, auth.ErrUserNotFound
	}
	return user, nil
}

func (m *memoryUserStore) GetUserByLogin(ctx context.Context, login string) (auth.User, error) {
	user, ok := m.usersByLogin[login]
	if !ok {
		return auth.User{}, auth.ErrUserNotFound
	}
	return user, nil
}

func (m *memoryUserStore) GetUserByID(ctx context.Context, id string) (auth.User, error) {
	user, ok := m.usersByID[id]
	if !ok {
		return auth.User{}, auth.ErrUserNotFound
	}
	return user, nil
}

func (m *memoryUserStore) UpdateUserProfile(ctx context.Context, id, login, email string) (auth.User, error) {
	user, ok := m.usersByID[id]
	if !ok {
		return auth.User{}, auth.ErrUserNotFound
	}
	user.Login = login
	user.Email = email
	user.UpdatedAt = time.Now().UTC()
	m.usersByID[id] = user
	m.usersByEmail[email] = user
	m.usersByLogin[login] = user
	return user, nil
}

type memoryKeyStore struct {
	keys map[string]auth.AuthKey
}

func newMemoryKeyStore() *memoryKeyStore {
	return &memoryKeyStore{keys: make(map[string]auth.AuthKey)}
}

func (m *memoryKeyStore) ListActiveKeys(ctx context.Context) ([]auth.AuthKey, error) {
	var active []auth.AuthKey
	for _, key := range m.keys {
		if key.IsActive {
			active = append(active, key)
		}
	}
	return active, nil
}

func (m *memoryKeyStore) GetKeyByID(ctx context.Context, keyID string) (auth.AuthKey, error) {
	key, ok := m.keys[keyID]
	if !ok {
		return auth.AuthKey{}, auth.ErrAuthKeyNotFound
	}
	return key, nil
}

func (m *memoryKeyStore) CreateAuthKey(ctx context.Context, key auth.AuthKey) (auth.AuthKey, error) {
	m.keys[key.KeyID] = key
	return key, nil
}

func newTokenManager(t *testing.T, store auth.KeyStore) *auth.TokenManager {
	t.Helper()
	manager, err := auth.NewTokenManager(context.Background(), store, time.Hour)
	if err != nil {
		t.Fatalf("failed to create token manager: %v", err)
	}
	return manager
}

func createUser(t *testing.T, store *memoryUserStore, login, email, password string) auth.User {
	t.Helper()
	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	user, err := store.CreateUser(context.Background(), login, email, hash)
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
	return user
}

func findAccessCookie(cookies []*http.Cookie) *http.Cookie {
	for _, cookie := range cookies {
		if cookie.Name == auth.AccessCookieName {
			return cookie
		}
	}
	return nil
}

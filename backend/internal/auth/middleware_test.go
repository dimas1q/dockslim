package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dimas1q/dockslim/backend/internal/citokens"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestMiddlewareUnauthorizedJSON(t *testing.T) {
	mw := NewMiddleware(nil, nil, nil)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)

	called := false
	handler := mw.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	handler.ServeHTTP(recorder, req)

	if called {
		t.Fatalf("expected handler to be blocked for missing auth header")
	}

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", recorder.Code)
	}

	if contentType := recorder.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("expected JSON content type, got %s", contentType)
	}

	expectedBody := "{\"error\":\"unauthorized\"}\n"
	if body := recorder.Body.String(); body != expectedBody {
		t.Fatalf("expected body %s, got %s", expectedBody, body)
	}
}

type tokenManagerStub struct {
	claims *AccessTokenClaims
	err    error
	called bool
}

func (t *tokenManagerStub) ValidateToken(ctx context.Context, tokenString string) (*AccessTokenClaims, error) {
	t.called = true
	return t.claims, t.err
}

type userStoreStub struct {
	user  User
	err   error
	calls int
}

func (u *userStoreStub) GetUserByID(ctx context.Context, id string) (User, error) {
	u.calls++
	return u.user, u.err
}

func (u *userStoreStub) CreateUser(ctx context.Context, login, email, passwordHash string) (User, error) {
	return User{}, nil
}

func (u *userStoreStub) GetUserByEmail(ctx context.Context, email string) (User, error) {
	return User{}, ErrUserNotFound
}

func (u *userStoreStub) GetUserByLogin(ctx context.Context, login string) (User, error) {
	return User{}, ErrUserNotFound
}

type ciAuthenticatorStub struct {
	token  citokens.Token
	err    error
	called bool
}

func (c *ciAuthenticatorStub) Authenticate(ctx context.Context, tokenString string) (citokens.Token, error) {
	c.called = true
	return c.token, c.err
}

func TestAuthenticateUsesCookieWhenNoAuthorizationHeader(t *testing.T) {
	userID := uuid.New()
	user := User{ID: userID, Email: "u@example.com"}
	tm := &tokenManagerStub{claims: &AccessTokenClaims{RegisteredClaims: jwt.RegisteredClaims{Subject: userID.String()}}}
	users := &userStoreStub{user: user}
	mw := NewMiddleware(tm, users, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "token"})
	rec := httptest.NewRecorder()

	called := false
	mw.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})).ServeHTTP(rec, req)

	if !called {
		t.Fatalf("expected handler to be called")
	}
	if !tm.called || users.calls != 1 {
		t.Fatalf("expected token and user lookup to be called")
	}
}

func TestAuthenticateRejectsCIToken(t *testing.T) {
	tm := &tokenManagerStub{}
	users := &userStoreStub{}
	mw := NewMiddleware(tm, users, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer "+citokens.TokenPrefix+"abc")
	rec := httptest.NewRecorder()

	mw.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("handler should not be called")
	})).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
	if tm.called {
		t.Fatalf("token manager should not be called for CI token")
	}
}

func TestAuthenticateUserOrCITokenAllowsCIToken(t *testing.T) {
	ciAuth := &ciAuthenticatorStub{token: citokens.Token{ID: uuid.New(), ProjectID: uuid.New()}}
	tm := &tokenManagerStub{}
	users := &userStoreStub{}
	mw := NewMiddleware(tm, users, ciAuth)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/ci/reports/image", nil)
	req.Header.Set("Authorization", "Bearer "+citokens.TokenPrefix+"abc")
	rec := httptest.NewRecorder()

	called := false
	mw.AuthenticateUserOrCIToken(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})).ServeHTTP(rec, req)

	if !called {
		t.Fatalf("expected handler to be called")
	}
	if !ciAuth.called {
		t.Fatalf("expected CI authenticator to be called")
	}
	if tm.called {
		t.Fatalf("token manager should not be used for CI token")
	}
}

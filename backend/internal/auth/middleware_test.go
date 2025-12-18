package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMiddlewareUnauthorizedJSON(t *testing.T) {
	mw := NewMiddleware(nil, nil)

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

	expectedBody := `{"error":"unauthorized"}`
	if body := recorder.Body.String(); body != expectedBody {
		t.Fatalf("expected body %s, got %s", expectedBody, body)
	}
}

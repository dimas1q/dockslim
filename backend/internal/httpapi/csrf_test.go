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
	"github.com/dimas1q/dockslim/backend/internal/projects"
	"github.com/google/uuid"
)

func TestCSRFProtectionRequiresHeader(t *testing.T) {
	userStore := newMemoryUserStore()
	tokenStore := newMemoryKeyStore()
	tokenManager := newTokenManager(t, tokenStore)
	authService := auth.NewService(userStore, tokenManager)
	projectRepo := newMemoryProjectRepo()
	projectService := projects.NewService(projectRepo)
	middleware := auth.NewMiddleware(tokenManager, userStore)
	authHandler := NewAuthHandler(authService, time.Hour, CookieConfig{
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})
	projectsHandler := NewProjectsHandler(projectService)

	router := NewRouter(Dependencies{
		AuthHandler:     authHandler,
		AuthMiddleware:  middleware,
		ProjectsHandler: projectsHandler,
		AllowedOrigins:  nil,
	})

	createUser(t, userStore, "csrf@example.com", "password123")

	loginRecorder := httptest.NewRecorder()
	loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(`{"email":"csrf@example.com","password":"password123"}`))
	router.ServeHTTP(loginRecorder, loginReq)

	if loginRecorder.Code != http.StatusOK {
		t.Fatalf("expected login status 200, got %d", loginRecorder.Code)
	}

	accessCookie := findCookie(loginRecorder.Result().Cookies(), auth.AccessCookieName)
	csrfCookie := findCookie(loginRecorder.Result().Cookies(), auth.CSRFCookieName)
	if accessCookie == nil || csrfCookie == nil {
		t.Fatalf("expected access and csrf cookies to be set")
	}

	payload := map[string]string{"name": "My Project"}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal project payload: %v", err)
	}

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects", bytes.NewBuffer(body))
	req.AddCookie(accessCookie)
	req.AddCookie(csrfCookie)
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", recorder.Code)
	}
	expected := "{\"error\":\"csrf validation failed\"}\n"
	if recorder.Body.String() != expected {
		t.Fatalf("expected body %s, got %s", expected, recorder.Body.String())
	}
}

func TestCSRFProtectionAcceptsMatchingHeader(t *testing.T) {
	userStore := newMemoryUserStore()
	tokenStore := newMemoryKeyStore()
	tokenManager := newTokenManager(t, tokenStore)
	authService := auth.NewService(userStore, tokenManager)
	projectRepo := newMemoryProjectRepo()
	projectService := projects.NewService(projectRepo)
	middleware := auth.NewMiddleware(tokenManager, userStore)
	authHandler := NewAuthHandler(authService, time.Hour, CookieConfig{
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})
	projectsHandler := NewProjectsHandler(projectService)

	router := NewRouter(Dependencies{
		AuthHandler:     authHandler,
		AuthMiddleware:  middleware,
		ProjectsHandler: projectsHandler,
		AllowedOrigins:  nil,
	})

	createUser(t, userStore, "csrf-success@example.com", "password123")

	loginRecorder := httptest.NewRecorder()
	loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(`{"email":"csrf-success@example.com","password":"password123"}`))
	router.ServeHTTP(loginRecorder, loginReq)

	accessCookie := findCookie(loginRecorder.Result().Cookies(), auth.AccessCookieName)
	csrfCookie := findCookie(loginRecorder.Result().Cookies(), auth.CSRFCookieName)
	if accessCookie == nil || csrfCookie == nil {
		t.Fatalf("expected access and csrf cookies to be set")
	}

	payload := map[string]string{"name": "My Project"}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal project payload: %v", err)
	}

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects", bytes.NewBuffer(body))
	req.AddCookie(accessCookie)
	req.AddCookie(csrfCookie)
	req.Header.Set(csrfHeaderName, csrfCookie.Value)
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", recorder.Code)
	}
}

type memoryProjectRepo struct {
	projects map[uuid.UUID]projects.Project
}

func newMemoryProjectRepo() *memoryProjectRepo {
	return &memoryProjectRepo{projects: make(map[uuid.UUID]projects.Project)}
}

func (m *memoryProjectRepo) CreateProjectWithOwner(ctx context.Context, name string, description *string, ownerID uuid.UUID) (projects.Project, error) {
	project := projects.Project{
		ID:        uuid.New(),
		Name:      name,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	if description != nil {
		project.Description = description
	}
	m.projects[project.ID] = project
	return project, nil
}

func (m *memoryProjectRepo) ListProjectsForUser(ctx context.Context, userID uuid.UUID) ([]projects.Project, error) {
	return nil, nil
}

func (m *memoryProjectRepo) GetProjectForUser(ctx context.Context, projectID, userID uuid.UUID) (projects.Project, error) {
	return projects.Project{}, projects.ErrProjectNotFound
}

func (m *memoryProjectRepo) GetMemberRole(ctx context.Context, projectID, userID uuid.UUID) (string, error) {
	return projects.RoleOwner, nil
}

func (m *memoryProjectRepo) UpdateProject(ctx context.Context, params projects.UpdateProjectParams) (projects.Project, error) {
	return projects.Project{}, projects.ErrProjectNotFound
}

func (m *memoryProjectRepo) DeleteProject(ctx context.Context, projectID uuid.UUID) error {
	return projects.ErrProjectNotFound
}

func findCookie(cookies []*http.Cookie, name string) *http.Cookie {
	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie
		}
	}
	return nil
}

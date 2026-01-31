package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dimas1q/dockslim/backend/internal/auth"
	"github.com/dimas1q/dockslim/backend/internal/projects"
	"github.com/google/uuid"
	"github.com/jackc/pgconn"
)

type projectRepoStub struct {
	createErr error
}

func (r *projectRepoStub) CreateProjectWithOwner(ctx context.Context, name string, description *string, ownerID uuid.UUID) (projects.Project, error) {
	if r.createErr != nil {
		return projects.Project{}, r.createErr
	}
	return projects.Project{ID: uuid.New(), Name: name}, nil
}

func (r *projectRepoStub) ListProjectsForUser(ctx context.Context, userID uuid.UUID) ([]projects.Project, error) {
	return nil, nil
}

func (r *projectRepoStub) GetProjectForUser(ctx context.Context, projectID, userID uuid.UUID) (projects.Project, error) {
	return projects.Project{}, projects.ErrProjectNotFound
}

func (r *projectRepoStub) GetMemberRole(ctx context.Context, projectID, userID uuid.UUID) (string, error) {
	return projects.RoleOwner, nil
}

func (r *projectRepoStub) UpdateProject(ctx context.Context, params projects.UpdateProjectParams) (projects.Project, error) {
	return projects.Project{}, projects.ErrProjectNotFound
}

func (r *projectRepoStub) DeleteProject(ctx context.Context, projectID uuid.UUID) error {
	return nil
}

func TestProjectsHandlerCreateUniqueViolationFallback(t *testing.T) {
	user := auth.User{ID: uuid.New()}
	repo := &projectRepoStub{createErr: &pgconn.PgError{Code: "23505"}}
	service := projects.NewService(repo)
	handler := NewProjectsHandler(service)

	payload := map[string]interface{}{"name": "Duplicate"}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects", bytes.NewBuffer(body))
	req = req.WithContext(auth.WithUser(req.Context(), user))

	rr := httptest.NewRecorder()
	handler.Create(rr, req)

	if rr.Code != http.StatusConflict {
		t.Fatalf("expected conflict 409, got %d", rr.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["error"] != "project with this name already exists" {
		t.Fatalf("unexpected error message: %s", resp["error"])
	}
}

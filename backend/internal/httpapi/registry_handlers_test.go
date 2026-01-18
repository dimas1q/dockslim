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
	"github.com/dimas1q/dockslim/backend/internal/registries"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type registryRepoStub struct {
	registries []registries.Registry
	createErr  error
}

func (r *registryRepoStub) ListRegistriesByProject(ctx context.Context, projectID uuid.UUID) ([]registries.Registry, error) {
	return r.registries, nil
}

func (r *registryRepoStub) CreateRegistry(ctx context.Context, params registries.CreateRegistryParams) (registries.Registry, error) {
	if r.createErr != nil {
		return registries.Registry{}, r.createErr
	}
	registry := registries.Registry{
		ID:          uuid.New(),
		ProjectID:   params.ProjectID,
		Name:        params.Name,
		Type:        params.Type,
		RegistryURL: params.RegistryURL,
		Username:    params.Username,
	}
	r.registries = append(r.registries, registry)
	return registry, nil
}

func (r *registryRepoStub) DeleteRegistry(ctx context.Context, projectID, registryID uuid.UUID) error {
	return nil
}

type membershipStub struct {
	role string
	err  error
}

func (m *membershipStub) GetMemberRole(ctx context.Context, projectID, userID uuid.UUID) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.role, nil
}

func TestRegistriesHandlerCreateOwnerOnly(t *testing.T) {
	projectID := uuid.New()
	user := auth.User{ID: uuid.New()}
	repo := &registryRepoStub{}
	members := &membershipStub{role: "member"}
	service := registries.NewService(repo, members, registries.EncryptionKey{KeyMaterial: []byte("01234567890123456789012345678901")})
	handler := NewRegistriesHandler(service)

	payload := map[string]string{
		"name":         "Registry",
		"type":         "generic",
		"registry_url": "https://registry.example.com",
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/"+projectID.String()+"/registries", bytes.NewBuffer(body))
	req = req.WithContext(auth.WithUser(req.Context(), user))
	req = withURLParam(req, "projectId", projectID.String())
	recorder := httptest.NewRecorder()

	handler.Create(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", recorder.Code)
	}
}

func TestRegistriesHandlerCreateDuplicateName(t *testing.T) {
	projectID := uuid.New()
	user := auth.User{ID: uuid.New()}
	repo := &registryRepoStub{createErr: registries.ErrRegistryNameConflict}
	members := &membershipStub{role: projects.RoleOwner}
	service := registries.NewService(repo, members, registries.EncryptionKey{KeyMaterial: []byte("01234567890123456789012345678901")})
	handler := NewRegistriesHandler(service)

	payload := map[string]string{
		"name":         "Registry",
		"type":         "generic",
		"registry_url": "https://registry.example.com",
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/"+projectID.String()+"/registries", bytes.NewBuffer(body))
	req = req.WithContext(auth.WithUser(req.Context(), user))
	req = withURLParam(req, "projectId", projectID.String())
	recorder := httptest.NewRecorder()

	handler.Create(recorder, req)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("expected status 409, got %d", recorder.Code)
	}
}

func TestRegistriesHandlerListMemberAllowed(t *testing.T) {
	projectID := uuid.New()
	user := auth.User{ID: uuid.New()}
	repo := &registryRepoStub{
		registries: []registries.Registry{
			{ID: uuid.New(), ProjectID: projectID, Name: "Registry", Type: "generic", RegistryURL: "https://registry.example.com"},
		},
	}
	members := &membershipStub{role: projects.RoleOwner}
	service := registries.NewService(repo, members, registries.EncryptionKey{KeyMaterial: []byte("01234567890123456789012345678901")})
	handler := NewRegistriesHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/"+projectID.String()+"/registries", nil)
	req = req.WithContext(auth.WithUser(req.Context(), user))
	req = withURLParam(req, "projectId", projectID.String())
	recorder := httptest.NewRecorder()

	handler.List(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
}

func TestRegistriesHandlerListNonMemberNotFound(t *testing.T) {
	projectID := uuid.New()
	user := auth.User{ID: uuid.New()}
	repo := &registryRepoStub{}
	members := &membershipStub{err: projects.ErrProjectMemberNotFound}
	service := registries.NewService(repo, members, registries.EncryptionKey{KeyMaterial: []byte("01234567890123456789012345678901")})
	handler := NewRegistriesHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/"+projectID.String()+"/registries", nil)
	req = req.WithContext(auth.WithUser(req.Context(), user))
	req = withURLParam(req, "projectId", projectID.String())
	recorder := httptest.NewRecorder()

	handler.List(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", recorder.Code)
	}
}

func withURLParam(r *http.Request, key, value string) *http.Request {
	routeContext := chi.NewRouteContext()
	routeContext.URLParams.Add(key, value)
	ctx := context.WithValue(r.Context(), chi.RouteCtxKey, routeContext)
	return r.WithContext(ctx)
}

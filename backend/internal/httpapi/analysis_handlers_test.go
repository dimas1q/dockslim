package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dimas1q/dockslim/backend/internal/analyses"
	"github.com/dimas1q/dockslim/backend/internal/auth"
	"github.com/dimas1q/dockslim/backend/internal/projects"
	"github.com/dimas1q/dockslim/backend/internal/registries"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type analysisRepoStub struct{}

func (r *analysisRepoStub) ListAnalysesByProject(ctx context.Context, projectID uuid.UUID) ([]analyses.ImageAnalysis, error) {
	return nil, nil
}

func (r *analysisRepoStub) GetAnalysisForProject(ctx context.Context, projectID, analysisID uuid.UUID) (analyses.ImageAnalysis, error) {
	return analyses.ImageAnalysis{}, analyses.ErrAnalysisNotFound
}

func (r *analysisRepoStub) CreateAnalysis(ctx context.Context, params analyses.CreateAnalysisParams) (analyses.ImageAnalysis, error) {
	return analyses.ImageAnalysis{
		ID:        uuid.New(),
		ProjectID: params.ProjectID,
		Image:     params.Image,
		Tag:       params.Tag,
		Status:    params.Status,
	}, nil
}

func (r *analysisRepoStub) DeleteAnalysis(ctx context.Context, projectID, analysisID uuid.UUID) error {
	return nil
}

func (r *analysisRepoStub) RerunAnalysis(ctx context.Context, projectID, analysisID uuid.UUID) error {
	return nil
}

type analysisRepoGetStub struct {
	analysis analyses.ImageAnalysis
}

func (r *analysisRepoGetStub) ListAnalysesByProject(ctx context.Context, projectID uuid.UUID) ([]analyses.ImageAnalysis, error) {
	return nil, nil
}

func (r *analysisRepoGetStub) GetAnalysisForProject(ctx context.Context, projectID, analysisID uuid.UUID) (analyses.ImageAnalysis, error) {
	return r.analysis, nil
}

func (r *analysisRepoGetStub) CreateAnalysis(ctx context.Context, params analyses.CreateAnalysisParams) (analyses.ImageAnalysis, error) {
	return analyses.ImageAnalysis{}, nil
}

func (r *analysisRepoGetStub) DeleteAnalysis(ctx context.Context, projectID, analysisID uuid.UUID) error {
	return nil
}

func (r *analysisRepoGetStub) RerunAnalysis(ctx context.Context, projectID, analysisID uuid.UUID) error {
	return nil
}

type registryStoreStub struct {
	err error
}

func (r *registryStoreStub) GetRegistryForProject(ctx context.Context, projectID, registryID uuid.UUID) (registries.Registry, error) {
	if r.err != nil {
		return registries.Registry{}, r.err
	}
	return registries.Registry{
		ID:          registryID,
		ProjectID:   projectID,
		RegistryURL: "https://registry.example.com",
	}, nil
}

type analysisMembershipStub struct {
	role string
	err  error
}

func (m *analysisMembershipStub) GetMemberRole(ctx context.Context, projectID, userID uuid.UUID) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.role, nil
}

func TestAnalysesHandlerCreateOwnerOnly(t *testing.T) {
	projectID := uuid.New()
	user := auth.User{ID: uuid.New()}
	repo := &analysisRepoStub{}
	members := &analysisMembershipStub{role: "member"}
	registryStore := &registryStoreStub{}
	service := analyses.NewService(repo, members, registryStore)
	handler := NewAnalysesHandler(service)

	payload := map[string]string{
		"registry_id": uuid.New().String(),
		"image":       "repo/image",
		"tag":         "latest",
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/"+projectID.String()+"/analyses", bytes.NewBuffer(body))
	req = req.WithContext(auth.WithUser(req.Context(), user))
	req = withURLParamAnalysis(req, "projectId", projectID.String())
	recorder := httptest.NewRecorder()

	handler.Create(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", recorder.Code)
	}
}

func TestAnalysesHandlerCreateRegistryMismatch(t *testing.T) {
	projectID := uuid.New()
	user := auth.User{ID: uuid.New()}
	repo := &analysisRepoStub{}
	members := &analysisMembershipStub{role: projects.RoleOwner}
	registryStore := &registryStoreStub{err: registries.ErrRegistryNotFound}
	service := analyses.NewService(repo, members, registryStore)
	handler := NewAnalysesHandler(service)

	payload := map[string]string{
		"registry_id": uuid.New().String(),
		"image":       "repo/image",
		"tag":         "latest",
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/"+projectID.String()+"/analyses", bytes.NewBuffer(body))
	req = req.WithContext(auth.WithUser(req.Context(), user))
	req = withURLParamAnalysis(req, "projectId", projectID.String())
	recorder := httptest.NewRecorder()

	handler.Create(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", recorder.Code)
	}
}

func TestAnalysesHandlerListMemberOnly(t *testing.T) {
	projectID := uuid.New()
	user := auth.User{ID: uuid.New()}
	repo := &analysisRepoStub{}
	members := &analysisMembershipStub{err: projects.ErrProjectMemberNotFound}
	registryStore := &registryStoreStub{}
	service := analyses.NewService(repo, members, registryStore)
	handler := NewAnalysesHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/"+projectID.String()+"/analyses", nil)
	req = req.WithContext(auth.WithUser(req.Context(), user))
	req = withURLParamAnalysis(req, "projectId", projectID.String())
	recorder := httptest.NewRecorder()

	handler.List(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", recorder.Code)
	}
}

func TestAnalysesHandlerGetIncludesRecommendations(t *testing.T) {
	projectID := uuid.New()
	analysisID := uuid.New()
	user := auth.User{ID: uuid.New()}
	resultJSON := json.RawMessage(`{"recommendations":[{"id":"large-image","severity":"warning","category":"size","title":"Large container image size","description":"The total image size exceeds 1 GB.","suggested_action":"Consider using a slimmer base image (alpine, distroless)."}]}`)
	repo := &analysisRepoGetStub{
		analysis: analyses.ImageAnalysis{
			ID:         analysisID,
			ProjectID:  projectID,
			Image:      "repo/image",
			Tag:        "latest",
			Status:     analyses.StatusCompleted,
			ResultJSON: resultJSON,
		},
	}
	members := &analysisMembershipStub{role: projects.RoleOwner}
	registryStore := &registryStoreStub{}
	service := analyses.NewService(repo, members, registryStore)
	handler := NewAnalysesHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/"+projectID.String()+"/analyses/"+analysisID.String(), nil)
	req = req.WithContext(auth.WithUser(req.Context(), user))
	req = withURLParamAnalysis(req, "projectId", projectID.String())
	req = withURLParamAnalysis(req, "analysisId", analysisID.String())
	recorder := httptest.NewRecorder()

	handler.Get(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	resultValue, ok := response["result_json"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected result_json object in response")
	}
	recommendations, ok := resultValue["recommendations"].([]interface{})
	if !ok || len(recommendations) == 0 {
		t.Fatalf("expected recommendations in result_json")
	}
	first, ok := recommendations[0].(map[string]interface{})
	if !ok || first["id"] != "large-image" {
		t.Fatalf("expected large-image recommendation")
	}
}

func withURLParamAnalysis(r *http.Request, key, value string) *http.Request {
	routeContext := chi.RouteContext(r.Context())
	if routeContext == nil {
		routeContext = chi.NewRouteContext()
	}
	routeContext.URLParams.Add(key, value)
	ctx := context.WithValue(r.Context(), chi.RouteCtxKey, routeContext)
	return r.WithContext(ctx)
}

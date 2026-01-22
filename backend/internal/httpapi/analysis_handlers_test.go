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

type analysisCompareRepoStub struct {
	analyses map[uuid.UUID]analyses.ImageAnalysis
}

func (r *analysisCompareRepoStub) ListAnalysesByProject(ctx context.Context, projectID uuid.UUID) ([]analyses.ImageAnalysis, error) {
	return nil, nil
}

func (r *analysisCompareRepoStub) GetAnalysisForProject(ctx context.Context, projectID, analysisID uuid.UUID) (analyses.ImageAnalysis, error) {
	analysis, ok := r.analyses[analysisID]
	if !ok {
		return analyses.ImageAnalysis{}, analyses.ErrAnalysisNotFound
	}
	return analysis, nil
}

func (r *analysisCompareRepoStub) CreateAnalysis(ctx context.Context, params analyses.CreateAnalysisParams) (analyses.ImageAnalysis, error) {
	return analyses.ImageAnalysis{}, nil
}

func (r *analysisCompareRepoStub) DeleteAnalysis(ctx context.Context, projectID, analysisID uuid.UUID) error {
	return nil
}

func (r *analysisCompareRepoStub) RerunAnalysis(ctx context.Context, projectID, analysisID uuid.UUID) error {
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

func TestAnalysesHandlerCompareMemberOnly(t *testing.T) {
	projectID := uuid.New()
	user := auth.User{ID: uuid.New()}
	repo := &analysisCompareRepoStub{analyses: map[uuid.UUID]analyses.ImageAnalysis{}}
	members := &analysisMembershipStub{err: projects.ErrProjectMemberNotFound}
	registryStore := &registryStoreStub{}
	service := analyses.NewService(repo, members, registryStore)
	handler := NewAnalysesHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/"+projectID.String()+"/analyses/compare?from="+uuid.New().String()+"&to="+uuid.New().String(), nil)
	req = req.WithContext(auth.WithUser(req.Context(), user))
	req = withURLParamAnalysis(req, "projectId", projectID.String())
	recorder := httptest.NewRecorder()

	handler.Compare(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", recorder.Code)
	}
}

func TestAnalysesHandlerCompareDifferentImages(t *testing.T) {
	projectID := uuid.New()
	user := auth.User{ID: uuid.New()}
	fromID := uuid.New()
	toID := uuid.New()
	repo := &analysisCompareRepoStub{
		analyses: map[uuid.UUID]analyses.ImageAnalysis{
			fromID: {
				ID:        fromID,
				ProjectID: projectID,
				Image:     "repo/app",
				Status:    analyses.StatusCompleted,
			},
			toID: {
				ID:        toID,
				ProjectID: projectID,
				Image:     "repo/other",
				Status:    analyses.StatusCompleted,
			},
		},
	}
	members := &analysisMembershipStub{role: projects.RoleOwner}
	registryStore := &registryStoreStub{}
	service := analyses.NewService(repo, members, registryStore)
	handler := NewAnalysesHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/"+projectID.String()+"/analyses/compare?from="+fromID.String()+"&to="+toID.String(), nil)
	req = req.WithContext(auth.WithUser(req.Context(), user))
	req = withURLParamAnalysis(req, "projectId", projectID.String())
	recorder := httptest.NewRecorder()

	handler.Compare(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", recorder.Code)
	}
}

func TestAnalysesHandlerCompareRequiresCompleted(t *testing.T) {
	projectID := uuid.New()
	user := auth.User{ID: uuid.New()}
	fromID := uuid.New()
	toID := uuid.New()
	repo := &analysisCompareRepoStub{
		analyses: map[uuid.UUID]analyses.ImageAnalysis{
			fromID: {
				ID:        fromID,
				ProjectID: projectID,
				Image:     "repo/app",
				Status:    analyses.StatusRunning,
			},
			toID: {
				ID:        toID,
				ProjectID: projectID,
				Image:     "repo/app",
				Status:    analyses.StatusCompleted,
			},
		},
	}
	members := &analysisMembershipStub{role: projects.RoleOwner}
	registryStore := &registryStoreStub{}
	service := analyses.NewService(repo, members, registryStore)
	handler := NewAnalysesHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/"+projectID.String()+"/analyses/compare?from="+fromID.String()+"&to="+toID.String(), nil)
	req = req.WithContext(auth.WithUser(req.Context(), user))
	req = withURLParamAnalysis(req, "projectId", projectID.String())
	recorder := httptest.NewRecorder()

	handler.Compare(recorder, req)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("expected status 409, got %d", recorder.Code)
	}
}

func TestAnalysesHandlerCompareDiff(t *testing.T) {
	projectID := uuid.New()
	user := auth.User{ID: uuid.New()}
	fromID := uuid.New()
	toID := uuid.New()
	fromResult := json.RawMessage(`{"layers":[{"digest":"sha256:aaa","size_bytes":10},{"digest":"sha256:bbb","size_bytes":30}],"total_size_bytes":40}`)
	toResult := json.RawMessage(`{"layers":[{"digest":"sha256:bbb","size_bytes":30},{"digest":"sha256:ccc","size_bytes":20}],"total_size_bytes":50}`)
	repo := &analysisCompareRepoStub{
		analyses: map[uuid.UUID]analyses.ImageAnalysis{
			fromID: {
				ID:             fromID,
				ProjectID:      projectID,
				Image:          "repo/app",
				Tag:            "1.0.0",
				Status:         analyses.StatusCompleted,
				TotalSizeBytes: func() *int64 { v := int64(40); return &v }(),
				ResultJSON:     fromResult,
			},
			toID: {
				ID:             toID,
				ProjectID:      projectID,
				Image:          "repo/app",
				Tag:            "1.1.0",
				Status:         analyses.StatusCompleted,
				TotalSizeBytes: func() *int64 { v := int64(50); return &v }(),
				ResultJSON:     toResult,
			},
		},
	}
	members := &analysisMembershipStub{role: projects.RoleOwner}
	registryStore := &registryStoreStub{}
	service := analyses.NewService(repo, members, registryStore)
	handler := NewAnalysesHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/"+projectID.String()+"/analyses/compare?from="+fromID.String()+"&to="+toID.String(), nil)
	req = req.WithContext(auth.WithUser(req.Context(), user))
	req = withURLParamAnalysis(req, "projectId", projectID.String())
	recorder := httptest.NewRecorder()

	handler.Compare(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	var response struct {
		Summary struct {
			TotalSizeDiffBytes int64 `json:"total_size_diff_bytes"`
			LayerCountDiff     int   `json:"layer_count_diff"`
		} `json:"summary"`
		Layers struct {
			Added []struct {
				Digest    string `json:"digest"`
				SizeBytes int64  `json:"size_bytes"`
			} `json:"added"`
			Removed []struct {
				Digest    string `json:"digest"`
				SizeBytes int64  `json:"size_bytes"`
			} `json:"removed"`
		} `json:"layers"`
	}
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if response.Summary.TotalSizeDiffBytes != 10 {
		t.Fatalf("expected size diff 10, got %d", response.Summary.TotalSizeDiffBytes)
	}
	if response.Summary.LayerCountDiff != 0 {
		t.Fatalf("expected layer count diff 0, got %d", response.Summary.LayerCountDiff)
	}
	if len(response.Layers.Added) != 1 || response.Layers.Added[0].Digest != "sha256:ccc" {
		t.Fatalf("expected added layer sha256:ccc")
	}
	if len(response.Layers.Removed) != 1 || response.Layers.Removed[0].Digest != "sha256:aaa" {
		t.Fatalf("expected removed layer sha256:aaa")
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

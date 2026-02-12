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
	"github.com/dimas1q/dockslim/backend/internal/budgets"
	"github.com/dimas1q/dockslim/backend/internal/featureflags"
	"github.com/dimas1q/dockslim/backend/internal/projects"
	"github.com/dimas1q/dockslim/backend/internal/registries"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type analysisRepoStub struct{}

func (r *analysisRepoStub) ListAnalysesByProject(ctx context.Context, projectID uuid.UUID) ([]analyses.ImageAnalysis, error) {
	return nil, nil
}

func (r *analysisRepoStub) ListHistory(ctx context.Context, projectID uuid.UUID, filter analyses.HistoryFilter) ([]analyses.HistoryItem, error) {
	return nil, nil
}

func (r *analysisRepoStub) ListTrends(ctx context.Context, projectID uuid.UUID, metric analyses.TrendMetric, filter analyses.HistoryFilter) ([]analyses.TrendPoint, error) {
	return nil, nil
}

func (r *analysisRepoStub) GetAnalysisForProject(ctx context.Context, projectID, analysisID uuid.UUID) (analyses.ImageAnalysis, error) {
	return analyses.ImageAnalysis{}, analyses.ErrAnalysisNotFound
}

func (r *analysisRepoStub) GetAnalysisByID(ctx context.Context, analysisID uuid.UUID) (analyses.ImageAnalysis, error) {
	return analyses.ImageAnalysis{}, analyses.ErrAnalysisNotFound
}

func (r *analysisRepoStub) GetLatestCompletedBaseline(ctx context.Context, projectID uuid.UUID, image, gitRef string, excludeID uuid.UUID) (analyses.ImageAnalysis, error) {
	return analyses.ImageAnalysis{}, analyses.ErrBaselineNotFound
}

func (r *analysisRepoStub) GetProjectPolicy(ctx context.Context, projectID uuid.UUID) (analyses.ProjectPolicy, error) {
	return analyses.ProjectPolicy{}, nil
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

func (r *analysisRepoGetStub) ListHistory(ctx context.Context, projectID uuid.UUID, filter analyses.HistoryFilter) ([]analyses.HistoryItem, error) {
	return nil, nil
}

func (r *analysisRepoGetStub) ListTrends(ctx context.Context, projectID uuid.UUID, metric analyses.TrendMetric, filter analyses.HistoryFilter) ([]analyses.TrendPoint, error) {
	return nil, nil
}

func (r *analysisRepoGetStub) GetAnalysisForProject(ctx context.Context, projectID, analysisID uuid.UUID) (analyses.ImageAnalysis, error) {
	return r.analysis, nil
}

func (r *analysisRepoGetStub) GetAnalysisByID(ctx context.Context, analysisID uuid.UUID) (analyses.ImageAnalysis, error) {
	return r.analysis, nil
}

func (r *analysisRepoGetStub) GetLatestCompletedBaseline(ctx context.Context, projectID uuid.UUID, image, gitRef string, excludeID uuid.UUID) (analyses.ImageAnalysis, error) {
	return analyses.ImageAnalysis{}, analyses.ErrBaselineNotFound
}

func (r *analysisRepoGetStub) GetProjectPolicy(ctx context.Context, projectID uuid.UUID) (analyses.ProjectPolicy, error) {
	return analyses.ProjectPolicy{}, nil
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

func (r *analysisCompareRepoStub) ListHistory(ctx context.Context, projectID uuid.UUID, filter analyses.HistoryFilter) ([]analyses.HistoryItem, error) {
	return nil, nil
}

func (r *analysisCompareRepoStub) ListTrends(ctx context.Context, projectID uuid.UUID, metric analyses.TrendMetric, filter analyses.HistoryFilter) ([]analyses.TrendPoint, error) {
	return nil, nil
}

func (r *analysisCompareRepoStub) GetAnalysisForProject(ctx context.Context, projectID, analysisID uuid.UUID) (analyses.ImageAnalysis, error) {
	analysis, ok := r.analyses[analysisID]
	if !ok {
		return analyses.ImageAnalysis{}, analyses.ErrAnalysisNotFound
	}
	return analysis, nil
}

func (r *analysisCompareRepoStub) GetAnalysisByID(ctx context.Context, analysisID uuid.UUID) (analyses.ImageAnalysis, error) {
	analysis, ok := r.analyses[analysisID]
	if !ok {
		return analyses.ImageAnalysis{}, analyses.ErrAnalysisNotFound
	}
	return analysis, nil
}

func (r *analysisCompareRepoStub) GetLatestCompletedBaseline(ctx context.Context, projectID uuid.UUID, image, gitRef string, excludeID uuid.UUID) (analyses.ImageAnalysis, error) {
	return analyses.ImageAnalysis{}, analyses.ErrBaselineNotFound
}

func (r *analysisCompareRepoStub) GetProjectPolicy(ctx context.Context, projectID uuid.UUID) (analyses.ProjectPolicy, error) {
	return analyses.ProjectPolicy{}, nil
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

type budgetResolverStub struct {
	resolved *budgets.ResolvedBudget
	err      error
}

type featureGateStub struct {
	featuresByUser map[uuid.UUID]featureflags.UserFeatures
}

func (b *budgetResolverStub) ResolveBudget(ctx context.Context, userID, projectID uuid.UUID, image string) (*budgets.ResolvedBudget, error) {
	return b.resolved, b.err
}

func (b *budgetResolverStub) ResolveBudgetForProject(ctx context.Context, projectID uuid.UUID, image string) (*budgets.ResolvedBudget, error) {
	return b.resolved, b.err
}

func (f *featureGateStub) GetUserFeatures(ctx context.Context, userID uuid.UUID) (featureflags.UserFeatures, error) {
	if f.featuresByUser == nil {
		return featureflags.UserFeatures{}, nil
	}
	if features, ok := f.featuresByUser[userID]; ok {
		return features, nil
	}
	return featureflags.UserFeatures{}, nil
}

func (f *featureGateStub) HasFeature(ctx context.Context, userID uuid.UUID, featureName string) (bool, error) {
	features, err := f.GetUserFeatures(ctx, userID)
	if err != nil {
		return false, err
	}
	value, ok := features.FeatureValue(featureName)
	if !ok {
		return false, nil
	}
	return featureflags.FeatureEnabled(value), nil
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
	service := analyses.NewService(repo, members, registryStore, &budgetResolverStub{})
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
	service := analyses.NewService(repo, members, registryStore, &budgetResolverStub{})
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
	service := analyses.NewService(repo, members, registryStore, &budgetResolverStub{})
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
	service := analyses.NewService(repo, members, registryStore, &budgetResolverStub{})
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
	service := analyses.NewService(repo, members, registryStore, &budgetResolverStub{})
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
	service := analyses.NewService(repo, members, registryStore, &budgetResolverStub{})
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
	service := analyses.NewService(repo, members, registryStore, &budgetResolverStub{})
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
	service := analyses.NewService(repo, members, registryStore, &budgetResolverStub{})
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

func TestAnalysesHandlerBaselineCompareNoBaseline(t *testing.T) {
	projectID := uuid.New()
	analysisID := uuid.New()
	user := auth.User{ID: uuid.New()}

	repo := &analysisRepoGetStub{
		analysis: analyses.ImageAnalysis{
			ID:        analysisID,
			ProjectID: projectID,
			Image:     "repo/image",
			Tag:       "latest",
			Status:    analyses.StatusCompleted,
		},
	}
	members := &analysisMembershipStub{role: projects.RoleOwner}
	registryStore := &registryStoreStub{}
	service := analyses.NewService(repo, members, registryStore, &budgetResolverStub{})
	handler := NewAnalysesHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/analyses/"+analysisID.String()+"/baseline-compare", nil)
	req = req.WithContext(auth.WithUser(req.Context(), user))
	req = withURLParamAnalysis(req, "analysisId", analysisID.String())
	recorder := httptest.NewRecorder()

	handler.BaselineCompare(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", recorder.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if response["error"] != "no baseline analysis found" {
		t.Fatalf("expected error message 'no baseline analysis found', got %q", response["error"])
	}
}

func TestAnalysesHandlerGetStripsRecommendationsForNonAdvancedPlan(t *testing.T) {
	projectID := uuid.New()
	analysisID := uuid.New()
	user := auth.User{ID: uuid.New()}
	resultJSON := json.RawMessage(`{
		"layers":[{"digest":"sha256:aaa","size_bytes":12}],
		"insights":{"warnings":["large layer","adv: premium-only warning"],"largest_layers":[1,2,3],"advanced_score":91},
		"recommendations":[{"id":"r1","title":"basic recommendation"},{"id":"adv_r2","title":"advanced recommendation","tier":"pro"}]
	}`)

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
	service := analyses.NewService(repo, members, registryStore, &budgetResolverStub{})
	flags := &featureGateStub{
		featuresByUser: map[uuid.UUID]featureflags.UserFeatures{
			user.ID: {
				Features: map[string]any{
					featureflags.FeatureAdvancedInsights: false,
				},
			},
		},
	}
	handler := NewAnalysesHandler(service, flags)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/"+projectID.String()+"/analyses/"+analysisID.String(), nil)
	req = req.WithContext(auth.WithUser(req.Context(), user))
	req = withURLParamAnalysis(req, "projectId", projectID.String())
	req = withURLParamAnalysis(req, "analysisId", analysisID.String())
	rec := httptest.NewRecorder()

	handler.Get(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var response map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	result, ok := response["result_json"].(map[string]any)
	if !ok {
		t.Fatalf("expected result_json object")
	}

	insights, ok := result["insights"].(map[string]any)
	if !ok {
		t.Fatalf("expected insights to remain for free plan")
	}
	if len(insights) != 1 {
		t.Fatalf("expected only warnings in insights, got %#v", insights)
	}
	warnings, ok := insights["warnings"].([]any)
	if !ok || len(warnings) != 1 {
		t.Fatalf("expected only basic warning to remain, got %#v", insights["warnings"])
	}
	if warnings[0] != "large layer" {
		t.Fatalf("expected basic warning to remain, got %#v", warnings[0])
	}
	if _, hasRecommendations := result["recommendations"]; hasRecommendations {
		t.Fatalf("expected recommendations to be hidden for non-advanced plan")
	}
}

func TestAnalysesHandlerBaselineCompareForbiddenWhenFeatureDisabled(t *testing.T) {
	user := auth.User{ID: uuid.New()}
	repo := &analysisRepoGetStub{}
	members := &analysisMembershipStub{role: projects.RoleOwner}
	registryStore := &registryStoreStub{}
	service := analyses.NewService(repo, members, registryStore, &budgetResolverStub{})
	flags := &featureGateStub{
		featuresByUser: map[uuid.UUID]featureflags.UserFeatures{
			user.ID: {
				Features: map[string]any{
					featureflags.FeatureBaselineSLA: false,
				},
			},
		},
	}
	handler := NewAnalysesHandler(service, flags)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/analyses/"+uuid.NewString()+"/baseline-compare", nil)
	req = req.WithContext(auth.WithUser(req.Context(), user))
	req = withURLParamAnalysis(req, "analysisId", uuid.NewString())
	rec := httptest.NewRecorder()

	handler.BaselineCompare(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestAnalysesHandlerExportJSONForbiddenWithoutFeature(t *testing.T) {
	projectID := uuid.New()
	analysisID := uuid.New()
	user := auth.User{ID: uuid.New()}
	repo := &analysisRepoGetStub{
		analysis: analyses.ImageAnalysis{
			ID:         analysisID,
			ProjectID:  projectID,
			Image:      "repo/image",
			Tag:        "latest",
			Status:     analyses.StatusCompleted,
			ResultJSON: json.RawMessage(`{"layers":[]}`),
		},
	}
	members := &analysisMembershipStub{role: projects.RoleOwner}
	registryStore := &registryStoreStub{}
	service := analyses.NewService(repo, members, registryStore, &budgetResolverStub{})
	flags := &featureGateStub{
		featuresByUser: map[uuid.UUID]featureflags.UserFeatures{
			user.ID: {Features: map[string]any{featureflags.FeatureExportJSON: false}},
		},
	}
	handler := NewAnalysesHandler(service, flags)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/"+projectID.String()+"/analyses/"+analysisID.String()+"/export/json", nil)
	req = req.WithContext(auth.WithUser(req.Context(), user))
	req = withURLParamAnalysis(req, "projectId", projectID.String())
	req = withURLParamAnalysis(req, "analysisId", analysisID.String())
	rec := httptest.NewRecorder()

	handler.ExportJSON(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestAnalysesHandlerExportJSONSanitizesNonAdvancedPayload(t *testing.T) {
	projectID := uuid.New()
	analysisID := uuid.New()
	user := auth.User{ID: uuid.New()}
	repo := &analysisRepoGetStub{
		analysis: analyses.ImageAnalysis{
			ID:        analysisID,
			ProjectID: projectID,
			Image:     "repo/image",
			Tag:       "latest",
			Status:    analyses.StatusCompleted,
			ResultJSON: json.RawMessage(`{
				"layers":[],
				"insights":{"warnings":["x","adv: hidden"],"layer_count":10,"advanced_score":99},
				"recommendations":[{"id":"r1"},{"id":"adv_r2","tier":"pro"}]
			}`),
		},
	}
	members := &analysisMembershipStub{role: projects.RoleOwner}
	registryStore := &registryStoreStub{}
	service := analyses.NewService(repo, members, registryStore, &budgetResolverStub{})
	flags := &featureGateStub{
		featuresByUser: map[uuid.UUID]featureflags.UserFeatures{
			user.ID: {
				Features: map[string]any{
					featureflags.FeatureExportJSON:       true,
					featureflags.FeatureAdvancedInsights: false,
				},
			},
		},
	}
	handler := NewAnalysesHandler(service, flags)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/"+projectID.String()+"/analyses/"+analysisID.String()+"/export/json", nil)
	req = req.WithContext(auth.WithUser(req.Context(), user))
	req = withURLParamAnalysis(req, "projectId", projectID.String())
	req = withURLParamAnalysis(req, "analysisId", analysisID.String())
	rec := httptest.NewRecorder()

	handler.ExportJSON(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode export payload: %v", err)
	}
	insights, ok := payload["insights"].(map[string]any)
	if !ok {
		t.Fatalf("expected insights object in export")
	}
	if len(insights) != 1 {
		t.Fatalf("expected only warnings in insights export, got %#v", insights)
	}
	warnings, ok := insights["warnings"].([]any)
	if !ok || len(warnings) != 1 {
		t.Fatalf("expected one basic warning in export, got %#v", insights["warnings"])
	}
	if warnings[0] != "x" {
		t.Fatalf("expected warning x in export, got %#v", warnings[0])
	}
	if _, hasRecommendations := payload["recommendations"]; hasRecommendations {
		t.Fatalf("expected recommendations key to be removed from export for non-advanced plan")
	}
}

func TestAnalysesHandlerExportPDFReturnsDocumentForUnicodeContent(t *testing.T) {
	projectID := uuid.New()
	analysisID := uuid.New()
	user := auth.User{ID: uuid.New()}
	repo := &analysisRepoGetStub{
		analysis: analyses.ImageAnalysis{
			ID:         analysisID,
			ProjectID:  projectID,
			Image:      "repo/пример",
			Tag:        "latest",
			Status:     analyses.StatusCompleted,
			ResultJSON: json.RawMessage(`{"layers":[]}`),
		},
	}
	members := &analysisMembershipStub{role: projects.RoleOwner}
	registryStore := &registryStoreStub{}
	service := analyses.NewService(repo, members, registryStore, &budgetResolverStub{})
	flags := &featureGateStub{
		featuresByUser: map[uuid.UUID]featureflags.UserFeatures{
			user.ID: {
				Features: map[string]any{
					featureflags.FeatureExportPDF: true,
				},
			},
		},
	}
	handler := NewAnalysesHandler(service, flags)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/"+projectID.String()+"/analyses/"+analysisID.String()+"/export/pdf", nil)
	req = req.WithContext(auth.WithUser(req.Context(), user))
	req = withURLParamAnalysis(req, "projectId", projectID.String())
	req = withURLParamAnalysis(req, "analysisId", analysisID.String())
	rec := httptest.NewRecorder()

	handler.ExportPDF(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/pdf" {
		t.Fatalf("expected Content-Type application/pdf, got %q", got)
	}
	if rec.Body.Len() == 0 {
		t.Fatalf("expected non-empty PDF body")
	}
	if !bytes.HasPrefix(rec.Body.Bytes(), []byte("%PDF")) {
		t.Fatalf("expected pdf signature in response body")
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

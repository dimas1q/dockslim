package analyses

import (
	"context"
	"testing"

	"github.com/dimas1q/dockslim/backend/internal/projects"
	"github.com/google/uuid"
)

type baselineRepoStub struct {
	analysis      ImageAnalysis
	baseline      ImageAnalysis
	policy        ProjectPolicy
	baselineErr   error
	lastRefBranch string
	lastExcludeID uuid.UUID
}

func (r *baselineRepoStub) ListAnalysesByProject(ctx context.Context, projectID uuid.UUID) ([]ImageAnalysis, error) {
	return nil, nil
}

func (r *baselineRepoStub) ListHistory(ctx context.Context, projectID uuid.UUID, filter HistoryFilter) ([]HistoryItem, error) {
	return nil, nil
}

func (r *baselineRepoStub) ListTrends(ctx context.Context, projectID uuid.UUID, metric TrendMetric, filter HistoryFilter) ([]TrendPoint, error) {
	return nil, nil
}

func (r *baselineRepoStub) GetAnalysisForProject(ctx context.Context, projectID, analysisID uuid.UUID) (ImageAnalysis, error) {
	return ImageAnalysis{}, ErrAnalysisNotFound
}

func (r *baselineRepoStub) GetAnalysisByID(ctx context.Context, analysisID uuid.UUID) (ImageAnalysis, error) {
	if r.analysis.ID == analysisID {
		return r.analysis, nil
	}
	return ImageAnalysis{}, ErrAnalysisNotFound
}

func (r *baselineRepoStub) GetLatestCompletedBaseline(ctx context.Context, projectID uuid.UUID, image, gitRef string, excludeID uuid.UUID) (ImageAnalysis, error) {
	r.lastRefBranch = gitRef
	r.lastExcludeID = excludeID
	if r.baselineErr != nil {
		return ImageAnalysis{}, r.baselineErr
	}
	return r.baseline, nil
}

func (r *baselineRepoStub) GetProjectPolicy(ctx context.Context, projectID uuid.UUID) (ProjectPolicy, error) {
	if r.policy.BaselineMode == "" && r.policy.BaselineRefBranch == "" {
		return defaultProjectPolicy(), nil
	}
	return r.policy, nil
}

func (r *baselineRepoStub) CreateAnalysis(ctx context.Context, params CreateAnalysisParams) (ImageAnalysis, error) {
	return ImageAnalysis{}, nil
}

func (r *baselineRepoStub) DeleteAnalysis(ctx context.Context, projectID, analysisID uuid.UUID) error {
	return nil
}

func (r *baselineRepoStub) RerunAnalysis(ctx context.Context, projectID, analysisID uuid.UUID) error {
	return nil
}

func TestBaselineCompareMainBranchExcludesSelf(t *testing.T) {
	projectID := uuid.New()
	analysisID := uuid.New()
	baselineID := uuid.New()
	userID := uuid.New()

	analysis := ImageAnalysis{
		ID:                analysisID,
		ProjectID:         projectID,
		Image:             "repo/app",
		Tag:               "v2",
		GitRef:            func() *string { v := "main"; return &v }(),
		Status:            StatusCompleted,
		TotalSizeBytes:    func() *int64 { v := int64(200); return &v }(),
		LayerCount:        func() *int { v := 5; return &v }(),
		LargestLayerBytes: func() *int64 { v := int64(120); return &v }(),
	}
	baseline := ImageAnalysis{
		ID:                baselineID,
		ProjectID:         projectID,
		Image:             "repo/app",
		Tag:               "v1",
		GitRef:            func() *string { v := "main"; return &v }(),
		Status:            StatusCompleted,
		TotalSizeBytes:    func() *int64 { v := int64(180); return &v }(),
		LayerCount:        func() *int { v := 4; return &v }(),
		LargestLayerBytes: func() *int64 { v := int64(100); return &v }(),
	}

	repo := &baselineRepoStub{analysis: analysis, baseline: baseline}
	members := &rerunMemberStub{role: projects.RoleOwner}
	service := NewService(repo, members, &rerunRegistryStub{}, &budgetResolverStub{})

	comparison, err := service.BaselineCompare(context.Background(), userID, analysisID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if comparison.Baseline.AnalysisID != baselineID {
		t.Fatalf("expected baseline %s, got %s", baselineID, comparison.Baseline.AnalysisID)
	}
	if repo.lastExcludeID != analysisID {
		t.Fatalf("expected exclude id %s, got %s", analysisID, repo.lastExcludeID)
	}
	if repo.lastRefBranch != "main" {
		t.Fatalf("expected ref branch main, got %s", repo.lastRefBranch)
	}
}

func TestBaselineCompareFeatureBranchUsesMain(t *testing.T) {
	projectID := uuid.New()
	analysisID := uuid.New()
	baselineID := uuid.New()
	userID := uuid.New()

	analysis := ImageAnalysis{
		ID:                analysisID,
		ProjectID:         projectID,
		Image:             "repo/app",
		Tag:               "feature",
		GitRef:            func() *string { v := "feature/login"; return &v }(),
		Status:            StatusCompleted,
		TotalSizeBytes:    func() *int64 { v := int64(210); return &v }(),
		LayerCount:        func() *int { v := 6; return &v }(),
		LargestLayerBytes: func() *int64 { v := int64(110); return &v }(),
	}
	baseline := ImageAnalysis{
		ID:                baselineID,
		ProjectID:         projectID,
		Image:             "repo/app",
		Tag:               "main",
		GitRef:            func() *string { v := "main"; return &v }(),
		Status:            StatusCompleted,
		TotalSizeBytes:    func() *int64 { v := int64(180); return &v }(),
		LayerCount:        func() *int { v := 5; return &v }(),
		LargestLayerBytes: func() *int64 { v := int64(100); return &v }(),
	}

	repo := &baselineRepoStub{analysis: analysis, baseline: baseline}
	members := &rerunMemberStub{role: projects.RoleOwner}
	service := NewService(repo, members, &rerunRegistryStub{}, &budgetResolverStub{})

	comparison, err := service.BaselineCompare(context.Background(), userID, analysisID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if comparison.Baseline.AnalysisID != baselineID {
		t.Fatalf("expected baseline %s, got %s", baselineID, comparison.Baseline.AnalysisID)
	}
	if repo.lastRefBranch != "main" {
		t.Fatalf("expected ref branch main, got %s", repo.lastRefBranch)
	}
}

func TestBaselineCompareMissingCurrentMetrics(t *testing.T) {
	projectID := uuid.New()
	analysisID := uuid.New()
	baselineID := uuid.New()
	userID := uuid.New()

	analysis := ImageAnalysis{
		ID:         analysisID,
		ProjectID:  projectID,
		Image:      "repo/app",
		Tag:        "v2",
		GitRef:     func() *string { v := "main"; return &v }(),
		Status:     StatusCompleted,
		LayerCount: func() *int { v := 5; return &v }(),
	}
	baseline := ImageAnalysis{
		ID:             baselineID,
		ProjectID:      projectID,
		Image:          "repo/app",
		Tag:            "v1",
		GitRef:         func() *string { v := "main"; return &v }(),
		Status:         StatusCompleted,
		TotalSizeBytes: func() *int64 { v := int64(180); return &v }(),
		LayerCount:     func() *int { v := 4; return &v }(),
	}

	repo := &baselineRepoStub{analysis: analysis, baseline: baseline}
	members := &rerunMemberStub{role: projects.RoleOwner}
	service := NewService(repo, members, &rerunRegistryStub{}, &budgetResolverStub{})

	_, err := service.BaselineCompare(context.Background(), userID, analysisID)
	if err != ErrBaselineMetricsUnavailable {
		t.Fatalf("expected ErrBaselineMetricsUnavailable, got %v", err)
	}
}

func TestBaselineCompareMissingBaselineMetrics(t *testing.T) {
	projectID := uuid.New()
	analysisID := uuid.New()
	baselineID := uuid.New()
	userID := uuid.New()

	analysis := ImageAnalysis{
		ID:             analysisID,
		ProjectID:      projectID,
		Image:          "repo/app",
		Tag:            "v2",
		GitRef:         func() *string { v := "main"; return &v }(),
		Status:         StatusCompleted,
		TotalSizeBytes: func() *int64 { v := int64(200); return &v }(),
		LayerCount:     func() *int { v := 5; return &v }(),
	}
	baseline := ImageAnalysis{
		ID:             baselineID,
		ProjectID:      projectID,
		Image:          "repo/app",
		Tag:            "v1",
		GitRef:         func() *string { v := "main"; return &v }(),
		Status:         StatusCompleted,
		TotalSizeBytes: func() *int64 { v := int64(180); return &v }(),
	}

	repo := &baselineRepoStub{analysis: analysis, baseline: baseline}
	members := &rerunMemberStub{role: projects.RoleOwner}
	service := NewService(repo, members, &rerunRegistryStub{}, &budgetResolverStub{})

	_, err := service.BaselineCompare(context.Background(), userID, analysisID)
	if err != ErrBaselineMetricsUnavailable {
		t.Fatalf("expected ErrBaselineMetricsUnavailable, got %v", err)
	}
}

package analyses

import (
	"context"
	"testing"

	"github.com/dimas1q/dockslim/backend/internal/budgets"
	"github.com/dimas1q/dockslim/backend/internal/projects"
	"github.com/dimas1q/dockslim/backend/internal/registries"
	"github.com/google/uuid"
)

type rerunRepoStub struct {
	analysis ImageAnalysis
	err      error
}

func (r *rerunRepoStub) ListAnalysesByProject(ctx context.Context, projectID uuid.UUID) ([]ImageAnalysis, error) {
	return nil, nil
}

func (r *rerunRepoStub) GetAnalysisForProject(ctx context.Context, projectID, analysisID uuid.UUID) (ImageAnalysis, error) {
	if r.err != nil {
		return ImageAnalysis{}, r.err
	}
	return r.analysis, nil
}

func (r *rerunRepoStub) CreateAnalysis(ctx context.Context, params CreateAnalysisParams) (ImageAnalysis, error) {
	return ImageAnalysis{}, nil
}

func (r *rerunRepoStub) DeleteAnalysis(ctx context.Context, projectID, analysisID uuid.UUID) error {
	return nil
}

func (r *rerunRepoStub) RerunAnalysis(ctx context.Context, projectID, analysisID uuid.UUID) error {
	return nil
}

type rerunMemberStub struct {
	role string
	err  error
}

func (m *rerunMemberStub) GetMemberRole(ctx context.Context, projectID, userID uuid.UUID) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.role, nil
}

type rerunRegistryStub struct{}

func (r *rerunRegistryStub) GetRegistryForProject(ctx context.Context, projectID, registryID uuid.UUID) (registries.Registry, error) {
	return registries.Registry{}, nil
}

type budgetResolverStub struct{}

func (b *budgetResolverStub) ResolveBudget(ctx context.Context, userID, projectID uuid.UUID, image string) (*budgets.ResolvedBudget, error) {
	return nil, nil
}

func (b *budgetResolverStub) ResolveBudgetForProject(ctx context.Context, projectID uuid.UUID, image string) (*budgets.ResolvedBudget, error) {
	return nil, nil
}

type compareRepoStub struct {
	items map[uuid.UUID]ImageAnalysis
}

func (r *compareRepoStub) ListAnalysesByProject(ctx context.Context, projectID uuid.UUID) ([]ImageAnalysis, error) {
	return nil, nil
}

func (r *compareRepoStub) GetAnalysisForProject(ctx context.Context, projectID, analysisID uuid.UUID) (ImageAnalysis, error) {
	item, ok := r.items[analysisID]
	if !ok {
		return ImageAnalysis{}, ErrAnalysisNotFound
	}
	return item, nil
}

func (r *compareRepoStub) CreateAnalysis(ctx context.Context, params CreateAnalysisParams) (ImageAnalysis, error) {
	return ImageAnalysis{}, nil
}

func (r *compareRepoStub) DeleteAnalysis(ctx context.Context, projectID, analysisID uuid.UUID) error {
	return nil
}

func (r *compareRepoStub) RerunAnalysis(ctx context.Context, projectID, analysisID uuid.UUID) error {
	return nil
}

func TestRerunAnalysisOwnerOnly(t *testing.T) {
	service := NewService(&rerunRepoStub{}, &rerunMemberStub{role: "member"}, &rerunRegistryStub{}, &budgetResolverStub{})
	err := service.RerunAnalysis(context.Background(), uuid.New(), uuid.New(), uuid.New())
	if err != ErrNotOwner {
		t.Fatalf("expected ErrNotOwner, got %v", err)
	}
}

func TestRerunAnalysisNonMemberReturnsNotFound(t *testing.T) {
	service := NewService(&rerunRepoStub{}, &rerunMemberStub{err: projects.ErrProjectMemberNotFound}, &rerunRegistryStub{}, &budgetResolverStub{})
	err := service.RerunAnalysis(context.Background(), uuid.New(), uuid.New(), uuid.New())
	if err != ErrProjectNotFound {
		t.Fatalf("expected ErrProjectNotFound, got %v", err)
	}
}

func TestCompareAnalysesOmitsBudgetWhenNoneConfigured(t *testing.T) {
	projectID := uuid.New()
	fromID := uuid.New()
	toID := uuid.New()
	analysisFrom := ImageAnalysis{
		ID:         fromID,
		ProjectID:  projectID,
		Image:      "repo/app",
		Status:     StatusCompleted,
		ResultJSON: nil,
	}
	analysisTo := ImageAnalysis{
		ID:         toID,
		ProjectID:  projectID,
		Image:      "repo/app",
		Status:     StatusCompleted,
		ResultJSON: nil,
	}
	repo := &compareRepoStub{items: map[uuid.UUID]ImageAnalysis{
		fromID: analysisFrom,
		toID:   analysisTo,
	}}
	members := &rerunMemberStub{role: "member"}
	service := NewService(repo, members, &rerunRegistryStub{}, &budgetResolverStub{})

	comparison, err := service.CompareAnalyses(context.Background(), uuid.New(), projectID, fromID, toID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if comparison.Budget != nil {
		t.Fatalf("expected budget to be nil when no budget configured")
	}
}

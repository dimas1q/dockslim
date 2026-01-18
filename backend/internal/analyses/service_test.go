package analyses

import (
	"context"
	"testing"

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

func TestRerunAnalysisOwnerOnly(t *testing.T) {
	service := NewService(&rerunRepoStub{}, &rerunMemberStub{role: "member"}, &rerunRegistryStub{})
	err := service.RerunAnalysis(context.Background(), uuid.New(), uuid.New(), uuid.New())
	if err != ErrNotOwner {
		t.Fatalf("expected ErrNotOwner, got %v", err)
	}
}

func TestRerunAnalysisNonMemberReturnsNotFound(t *testing.T) {
	service := NewService(&rerunRepoStub{}, &rerunMemberStub{err: projects.ErrProjectMemberNotFound}, &rerunRegistryStub{})
	err := service.RerunAnalysis(context.Background(), uuid.New(), uuid.New(), uuid.New())
	if err != ErrProjectNotFound {
		t.Fatalf("expected ErrProjectNotFound, got %v", err)
	}
}

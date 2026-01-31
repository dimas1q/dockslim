package budgets

import (
	"context"
	"testing"

	"github.com/dimas1q/dockslim/backend/internal/projects"
	"github.com/google/uuid"
)

type repoStub struct {
	budgets          []Budget
	resolved         *ResolvedBudget
	lastProjectID    uuid.UUID
	lastImage        string
	updatedThreshold ResolvedBudget
}

func (r *repoStub) ListBudgetsByProject(ctx context.Context, projectID uuid.UUID) ([]Budget, error) {
	r.lastProjectID = projectID
	return r.budgets, nil
}

func (r *repoStub) UpsertDefaultBudget(ctx context.Context, projectID uuid.UUID, thresholds ResolvedBudget) (Budget, error) {
	r.lastProjectID = projectID
	r.updatedThreshold = thresholds
	return Budget{ID: uuid.New(), ProjectID: projectID}, nil
}

func (r *repoStub) CreateBudgetOverride(ctx context.Context, projectID uuid.UUID, image string, thresholds ResolvedBudget) (Budget, error) {
	r.lastProjectID = projectID
	r.lastImage = image
	r.updatedThreshold = thresholds
	return Budget{ID: uuid.New(), ProjectID: projectID, Image: &image}, nil
}

func (r *repoStub) UpdateBudget(ctx context.Context, budgetID, projectID uuid.UUID, image *string, thresholds ResolvedBudget) (Budget, error) {
	r.lastProjectID = projectID
	if image != nil {
		r.lastImage = *image
	}
	r.updatedThreshold = thresholds
	return Budget{ID: budgetID, ProjectID: projectID, Image: image}, nil
}

func (r *repoStub) DeleteBudget(ctx context.Context, budgetID, projectID uuid.UUID) error {
	r.lastProjectID = projectID
	return nil
}

func (r *repoStub) ResolveBudgetForImage(ctx context.Context, projectID uuid.UUID, image string) (*ResolvedBudget, error) {
	r.lastProjectID = projectID
	r.lastImage = image
	return r.resolved, nil
}

type memberStub struct {
	role string
	err  error
}

func (m *memberStub) GetMemberRole(ctx context.Context, projectID, userID uuid.UUID) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.role, nil
}

func TestResolveBudgetOverrideWins(t *testing.T) {
	repo := &repoStub{resolved: &ResolvedBudget{FailDeltaBytes: ptr(10)}}
	members := &memberStub{role: projects.RoleOwner}
	svc := NewService(repo, members)

	projectID := uuid.New()
	userID := uuid.New()
	image := "company/app"

	res, err := svc.ResolveBudget(context.Background(), userID, projectID, image)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res == nil || res.FailDeltaBytes == nil || *res.FailDeltaBytes != 10 {
		t.Fatalf("expected fail delta 10")
	}
	if repo.lastImage != image {
		t.Fatalf("expected resolve called with image")
	}
}

func TestCreateOverrideValidatesImage(t *testing.T) {
	repo := &repoStub{}
	members := &memberStub{role: projects.RoleOwner}
	svc := NewService(repo, members)

	_, err := svc.CreateOverride(context.Background(), uuid.New(), uuid.New(), OverrideBudgetInput{Image: "  "})
	if err != ErrInvalidImage {
		t.Fatalf("expected ErrInvalidImage, got %v", err)
	}
}

func TestUpsertDefaultValidatesThresholds(t *testing.T) {
	repo := &repoStub{}
	members := &memberStub{role: projects.RoleOwner}
	svc := NewService(repo, members)

	invalid := int64(-1)
	_, err := svc.UpsertDefault(context.Background(), uuid.New(), uuid.New(), DefaultBudgetInput{Thresholds: ThresholdsInput{HardLimitBytes: &invalid}})
	if err != ErrInvalidThreshold {
		t.Fatalf("expected ErrInvalidThreshold, got %v", err)
	}
}

func TestNonMemberResolveReturnsNotFound(t *testing.T) {
	repo := &repoStub{}
	members := &memberStub{err: projects.ErrProjectMemberNotFound}
	svc := NewService(repo, members)

	_, err := svc.ResolveBudget(context.Background(), uuid.New(), uuid.New(), "img")
	if err != ErrProjectNotFound {
		t.Fatalf("expected ErrProjectNotFound, got %v", err)
	}
}

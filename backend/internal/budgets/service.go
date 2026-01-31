package budgets

import (
	"context"
	"errors"
	"math"
	"strings"

	"github.com/dimas1q/dockslim/backend/internal/projects"
	"github.com/google/uuid"
)

var (
	ErrProjectNotFound    = errors.New("project not found")
	ErrNotOwner           = errors.New("user is not project owner")
	ErrInvalidImage       = errors.New("invalid image")
	ErrInvalidThreshold   = errors.New("invalid threshold")
	ErrInvalidBudgetPatch = errors.New("invalid budget patch")
)

type RepositoryStore interface {
	ListBudgetsByProject(ctx context.Context, projectID uuid.UUID) ([]Budget, error)
	UpsertDefaultBudget(ctx context.Context, projectID uuid.UUID, thresholds ResolvedBudget) (Budget, error)
	CreateBudgetOverride(ctx context.Context, projectID uuid.UUID, image string, thresholds ResolvedBudget) (Budget, error)
	UpdateBudget(ctx context.Context, budgetID, projectID uuid.UUID, image *string, thresholds ResolvedBudget) (Budget, error)
	DeleteBudget(ctx context.Context, budgetID, projectID uuid.UUID) error
	ResolveBudgetForImage(ctx context.Context, projectID uuid.UUID, image string) (*ResolvedBudget, error)
}

type MembershipStore interface {
	GetMemberRole(ctx context.Context, projectID, userID uuid.UUID) (string, error)
}

type Service struct {
	repo    RepositoryStore
	members MembershipStore
}

func NewService(repo RepositoryStore, members MembershipStore) *Service {
	return &Service{repo: repo, members: members}
}

type ThresholdsInput struct {
	WarnDeltaBytes *int64
	FailDeltaBytes *int64
	HardLimitBytes *int64
}

type DefaultBudgetInput struct {
	Thresholds ThresholdsInput
}

type OverrideBudgetInput struct {
	Image      string
	Thresholds ThresholdsInput
}

type UpdateBudgetInput struct {
	Image      *string
	Thresholds ThresholdsInput
}

func (s *Service) GetBudgets(ctx context.Context, userID, projectID uuid.UUID) ([]Budget, error) {
	_, err := s.members.GetMemberRole(ctx, projectID, userID)
	if err != nil {
		if errors.Is(err, projects.ErrProjectMemberNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, err
	}

	return s.repo.ListBudgetsByProject(ctx, projectID)
}

func (s *Service) UpsertDefault(ctx context.Context, userID, projectID uuid.UUID, input DefaultBudgetInput) (Budget, error) {
	if err := s.ensureOwner(ctx, userID, projectID); err != nil {
		return Budget{}, err
	}

	thresholds, err := normalizeThresholds(input.Thresholds)
	if err != nil {
		return Budget{}, err
	}

	return s.repo.UpsertDefaultBudget(ctx, projectID, thresholds)
}

func (s *Service) CreateOverride(ctx context.Context, userID, projectID uuid.UUID, input OverrideBudgetInput) (Budget, error) {
	if err := s.ensureOwner(ctx, userID, projectID); err != nil {
		return Budget{}, err
	}

	image := strings.TrimSpace(input.Image)
	if image == "" {
		return Budget{}, ErrInvalidImage
	}

	thresholds, err := normalizeThresholds(input.Thresholds)
	if err != nil {
		return Budget{}, err
	}

	budget, err := s.repo.CreateBudgetOverride(ctx, projectID, image, thresholds)
	if err != nil {
		if errors.Is(err, ErrBudgetConflict) {
			return Budget{}, ErrBudgetConflict
		}
	}

	return budget, err
}

func (s *Service) UpdateBudget(ctx context.Context, userID, projectID, budgetID uuid.UUID, input UpdateBudgetInput) (Budget, error) {
	if err := s.ensureOwner(ctx, userID, projectID); err != nil {
		return Budget{}, err
	}

	if input.Image == nil && input.Thresholds.WarnDeltaBytes == nil && input.Thresholds.FailDeltaBytes == nil && input.Thresholds.HardLimitBytes == nil {
		return Budget{}, ErrInvalidBudgetPatch
	}

	var image *string
	if input.Image != nil {
		clean := strings.TrimSpace(*input.Image)
		if clean == "" {
			return Budget{}, ErrInvalidImage
		}
		image = &clean
	}

	thresholds, err := normalizeThresholds(input.Thresholds)
	if err != nil {
		return Budget{}, err
	}

	return s.repo.UpdateBudget(ctx, budgetID, projectID, image, thresholds)
}

func (s *Service) DeleteBudget(ctx context.Context, userID, projectID, budgetID uuid.UUID) error {
	if err := s.ensureOwner(ctx, userID, projectID); err != nil {
		return err
	}
	return s.repo.DeleteBudget(ctx, budgetID, projectID)
}

func (s *Service) ResolveBudget(ctx context.Context, userID, projectID uuid.UUID, image string) (*ResolvedBudget, error) {
	_, err := s.members.GetMemberRole(ctx, projectID, userID)
	if err != nil {
		if errors.Is(err, projects.ErrProjectMemberNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, err
	}
	return s.repo.ResolveBudgetForImage(ctx, projectID, image)
}

func (s *Service) ensureOwner(ctx context.Context, userID, projectID uuid.UUID) error {
	role, err := s.members.GetMemberRole(ctx, projectID, userID)
	if err != nil {
		if errors.Is(err, projects.ErrProjectMemberNotFound) {
			return ErrProjectNotFound
		}
		return err
	}
	if role != projects.RoleOwner {
		return ErrNotOwner
	}
	return nil
}

func normalizeThresholds(input ThresholdsInput) (ResolvedBudget, error) {
	validate := func(v *int64) (*int64, error) {
		if v == nil {
			return nil, nil
		}
		if *v < 0 {
			return nil, ErrInvalidThreshold
		}
		return v, nil
	}

	warn, err := validate(input.WarnDeltaBytes)
	if err != nil {
		return ResolvedBudget{}, err
	}
	fail, err := validate(input.FailDeltaBytes)
	if err != nil {
		return ResolvedBudget{}, err
	}
	hard, err := validate(input.HardLimitBytes)
	if err != nil {
		return ResolvedBudget{}, err
	}

	return ResolvedBudget{
		WarnDeltaBytes: warn,
		FailDeltaBytes: fail,
		HardLimitBytes: hard,
	}, nil
}

// MBToBytes converts MiB value to bytes with overflow protection.
func MBToBytes(mb *int64) (*int64, error) {
	if mb == nil {
		return nil, nil
	}
	if *mb < 0 {
		return nil, ErrInvalidThreshold
	}
	const factor = int64(1024 * 1024)
	if *mb > math.MaxInt64/factor {
		return nil, ErrInvalidThreshold
	}
	value := *mb * factor
	return &value, nil
}

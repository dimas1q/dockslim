package analyses

import (
	"context"
	"errors"
	"strings"

	"github.com/dimas1q/dockslim/backend/internal/budgets"
	"github.com/dimas1q/dockslim/backend/internal/projects"
	"github.com/dimas1q/dockslim/backend/internal/registries"
	"github.com/google/uuid"
)

var (
	ErrProjectNotFound        = errors.New("project not found")
	ErrNotOwner               = errors.New("user is not project owner")
	ErrInvalidImage           = errors.New("invalid image")
	ErrInvalidRegistry        = errors.New("invalid registry")
	ErrRegistryMismatch       = errors.New("image registry does not match selected registry")
	ErrAnalysisRunning        = errors.New("analysis is running")
	ErrAnalysesDifferentImage = errors.New("analyses must be for the same image")
	ErrAnalysesNotCompleted   = errors.New("both analyses must be completed")
)

type RepositoryStore interface {
	ListAnalysesByProject(ctx context.Context, projectID uuid.UUID) ([]ImageAnalysis, error)
	GetAnalysisForProject(ctx context.Context, projectID, analysisID uuid.UUID) (ImageAnalysis, error)
	CreateAnalysis(ctx context.Context, params CreateAnalysisParams) (ImageAnalysis, error)
	DeleteAnalysis(ctx context.Context, projectID, analysisID uuid.UUID) error
	RerunAnalysis(ctx context.Context, projectID, analysisID uuid.UUID) error
}

type MembershipStore interface {
	GetMemberRole(ctx context.Context, projectID, userID uuid.UUID) (string, error)
}

type RegistryStore interface {
	GetRegistryForProject(ctx context.Context, projectID, registryID uuid.UUID) (registries.Registry, error)
}

type Service struct {
	repo       RepositoryStore
	members    MembershipStore
	registries RegistryStore
	budgets    BudgetResolver
}

type BudgetResolver interface {
	ResolveBudget(ctx context.Context, userID, projectID uuid.UUID, image string) (*budgets.ResolvedBudget, error)
}

func NewService(repo RepositoryStore, members MembershipStore, registries RegistryStore, budgets BudgetResolver) *Service {
	return &Service{repo: repo, members: members, registries: registries, budgets: budgets}
}

func (s *Service) ListAnalyses(ctx context.Context, userID, projectID uuid.UUID) ([]ImageAnalysis, error) {
	_, err := s.members.GetMemberRole(ctx, projectID, userID)
	if err != nil {
		if errors.Is(err, projects.ErrProjectMemberNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, err
	}
	return s.repo.ListAnalysesByProject(ctx, projectID)
}

func (s *Service) GetAnalysis(ctx context.Context, userID, projectID, analysisID uuid.UUID) (ImageAnalysis, error) {
	_, err := s.members.GetMemberRole(ctx, projectID, userID)
	if err != nil {
		if errors.Is(err, projects.ErrProjectMemberNotFound) {
			return ImageAnalysis{}, ErrProjectNotFound
		}
		return ImageAnalysis{}, err
	}
	return s.repo.GetAnalysisForProject(ctx, projectID, analysisID)
}

func (s *Service) CreateAnalysis(ctx context.Context, userID, projectID uuid.UUID, registryID uuid.UUID, image, tag string) (ImageAnalysis, error) {
	role, err := s.members.GetMemberRole(ctx, projectID, userID)
	if err != nil {
		if errors.Is(err, projects.ErrProjectMemberNotFound) {
			return ImageAnalysis{}, ErrProjectNotFound
		}
		return ImageAnalysis{}, err
	}
	if role != projects.RoleOwner {
		return ImageAnalysis{}, ErrNotOwner
	}

	if registryID == uuid.Nil {
		return ImageAnalysis{}, ErrInvalidRegistry
	}

	cleanImage := strings.TrimSpace(image)
	if cleanImage == "" {
		return ImageAnalysis{}, ErrInvalidImage
	}

	cleanTag := strings.TrimSpace(tag)
	if cleanTag == "" {
		cleanTag = "latest"
	}

	registry, err := s.registries.GetRegistryForProject(ctx, projectID, registryID)
	if err != nil {
		if errors.Is(err, registries.ErrRegistryNotFound) {
			return ImageAnalysis{}, err
		}
		return ImageAnalysis{}, err
	}

	normalizedImage, err := normalizeImageReference(cleanImage, registry.RegistryURL)
	if err != nil {
		return ImageAnalysis{}, err
	}

	return s.repo.CreateAnalysis(ctx, CreateAnalysisParams{
		ProjectID:  projectID,
		RegistryID: &registryID,
		Image:      normalizedImage,
		Tag:        cleanTag,
		Status:     StatusQueued,
	})
}

func (s *Service) DeleteAnalysis(ctx context.Context, userID, projectID, analysisID uuid.UUID) error {
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

	return s.repo.DeleteAnalysis(ctx, projectID, analysisID)
}

func (s *Service) RerunAnalysis(ctx context.Context, userID, projectID, analysisID uuid.UUID) error {
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

	analysis, err := s.repo.GetAnalysisForProject(ctx, projectID, analysisID)
	if err != nil {
		return err
	}
	if analysis.Status == StatusRunning {
		return ErrAnalysisRunning
	}

	return s.repo.RerunAnalysis(ctx, projectID, analysisID)
}

func (s *Service) CompareAnalyses(ctx context.Context, userID, projectID, fromID, toID uuid.UUID) (Comparison, error) {
	_, err := s.members.GetMemberRole(ctx, projectID, userID)
	if err != nil {
		if errors.Is(err, projects.ErrProjectMemberNotFound) {
			return Comparison{}, ErrProjectNotFound
		}
		return Comparison{}, err
	}

	fromAnalysis, err := s.repo.GetAnalysisForProject(ctx, projectID, fromID)
	if err != nil {
		return Comparison{}, err
	}
	toAnalysis, err := s.repo.GetAnalysisForProject(ctx, projectID, toID)
	if err != nil {
		return Comparison{}, err
	}

	if fromAnalysis.Image != toAnalysis.Image {
		return Comparison{}, ErrAnalysesDifferentImage
	}
	if fromAnalysis.Status != StatusCompleted || toAnalysis.Status != StatusCompleted {
		return Comparison{}, ErrAnalysesNotCompleted
	}

	comparison, err := BuildComparison(fromAnalysis, toAnalysis)
	if err != nil {
		return Comparison{}, err
	}

	if s.budgets != nil {
		resolved, err := s.budgets.ResolveBudget(ctx, userID, projectID, comparison.Image)
		if err != nil {
			return Comparison{}, err
		}
		if resolved != nil {
			eval := budgets.EvaluateBudget(comparison.From.TotalSizeBytes, comparison.To.TotalSizeBytes, resolved)
			comparison.Budget = &eval
		}
	}

	return comparison, nil
}

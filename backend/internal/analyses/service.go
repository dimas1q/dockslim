package analyses

import (
	"context"
	"errors"
	"strings"

	"github.com/dimas1q/dockslim/backend/internal/projects"
	"github.com/dimas1q/dockslim/backend/internal/registries"
	"github.com/google/uuid"
)

var (
	ErrProjectNotFound = errors.New("project not found")
	ErrNotOwner        = errors.New("user is not project owner")
	ErrInvalidImage    = errors.New("invalid image")
	ErrInvalidRegistry = errors.New("invalid registry")
)

type RepositoryStore interface {
	ListAnalysesByProject(ctx context.Context, projectID uuid.UUID) ([]ImageAnalysis, error)
	GetAnalysisForProject(ctx context.Context, projectID, analysisID uuid.UUID) (ImageAnalysis, error)
	CreateAnalysis(ctx context.Context, params CreateAnalysisParams) (ImageAnalysis, error)
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
}

func NewService(repo RepositoryStore, members MembershipStore, registries RegistryStore) *Service {
	return &Service{repo: repo, members: members, registries: registries}
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

	if _, err := s.registries.GetRegistryForProject(ctx, projectID, registryID); err != nil {
		if errors.Is(err, registries.ErrRegistryNotFound) {
			return ImageAnalysis{}, err
		}
		return ImageAnalysis{}, err
	}

	return s.repo.CreateAnalysis(ctx, CreateAnalysisParams{
		ProjectID:  projectID,
		RegistryID: &registryID,
		Image:      cleanImage,
		Tag:        cleanTag,
		Status:     StatusQueued,
	})
}

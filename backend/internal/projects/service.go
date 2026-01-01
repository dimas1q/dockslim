package projects

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrInvalidProjectName = errors.New("invalid project name")
	ErrNotOwner           = errors.New("user is not project owner")
)

type RepositoryStore interface {
	CreateProjectWithOwner(ctx context.Context, name string, ownerID uuid.UUID) (Project, error)
	ListProjectsForUser(ctx context.Context, userID uuid.UUID) ([]Project, error)
	GetProjectForUser(ctx context.Context, projectID, userID uuid.UUID) (Project, error)
	GetMemberRole(ctx context.Context, projectID, userID uuid.UUID) (string, error)
	UpdateProjectName(ctx context.Context, projectID uuid.UUID, name string) (Project, error)
	DeleteProject(ctx context.Context, projectID uuid.UUID) error
}

type Service struct {
	repo RepositoryStore
}

func NewService(repo RepositoryStore) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateProject(ctx context.Context, ownerID uuid.UUID, name string) (Project, error) {
	cleanName, err := validateProjectName(name)
	if err != nil {
		return Project{}, err
	}

	return s.repo.CreateProjectWithOwner(ctx, cleanName, ownerID)
}

func (s *Service) ListProjects(ctx context.Context, userID uuid.UUID) ([]Project, error) {
	return s.repo.ListProjectsForUser(ctx, userID)
}

func (s *Service) GetProject(ctx context.Context, userID, projectID uuid.UUID) (Project, error) {
	return s.repo.GetProjectForUser(ctx, projectID, userID)
}

func (s *Service) UpdateProjectName(ctx context.Context, userID, projectID uuid.UUID, name string) (Project, error) {
	cleanName, err := validateProjectName(name)
	if err != nil {
		return Project{}, err
	}

	role, err := s.repo.GetMemberRole(ctx, projectID, userID)
	if err != nil {
		if errors.Is(err, ErrProjectMemberNotFound) {
			return Project{}, ErrProjectNotFound
		}
		return Project{}, err
	}
	if role != RoleOwner {
		return Project{}, ErrNotOwner
	}

	return s.repo.UpdateProjectName(ctx, projectID, cleanName)
}

func (s *Service) DeleteProject(ctx context.Context, userID, projectID uuid.UUID) error {
	role, err := s.repo.GetMemberRole(ctx, projectID, userID)
	if err != nil {
		if errors.Is(err, ErrProjectMemberNotFound) {
			return ErrProjectNotFound
		}
		return err
	}
	if role != RoleOwner {
		return ErrNotOwner
	}

	return s.repo.DeleteProject(ctx, projectID)
}

func (s *Service) GetMemberRole(ctx context.Context, userID, projectID uuid.UUID) (string, error) {
	role, err := s.repo.GetMemberRole(ctx, projectID, userID)
	if err != nil {
		if errors.Is(err, ErrProjectMemberNotFound) {
			return "", ErrProjectNotFound
		}
		return "", err
	}
	return role, nil
}

func validateProjectName(name string) (string, error) {
	clean := strings.TrimSpace(name)
	if len(clean) < 2 || len(clean) > 100 {
		return "", ErrInvalidProjectName
	}
	return clean, nil
}

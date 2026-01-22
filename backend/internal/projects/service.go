package projects

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrInvalidProjectName  = errors.New("invalid project name")
	ErrInvalidProjectPatch = errors.New("invalid project patch")
	ErrNotOwner            = errors.New("user is not project owner")
)

type RepositoryStore interface {
	CreateProjectWithOwner(ctx context.Context, name string, description *string, ownerID uuid.UUID) (Project, error)
	ListProjectsForUser(ctx context.Context, userID uuid.UUID) ([]Project, error)
	GetProjectForUser(ctx context.Context, projectID, userID uuid.UUID) (Project, error)
	GetMemberRole(ctx context.Context, projectID, userID uuid.UUID) (string, error)
	UpdateProject(ctx context.Context, params UpdateProjectParams) (Project, error)
	DeleteProject(ctx context.Context, projectID uuid.UUID) error
}

type Service struct {
	repo RepositoryStore
}

func NewService(repo RepositoryStore) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateProject(ctx context.Context, ownerID uuid.UUID, name string, description *string) (Project, error) {
	cleanName, err := validateProjectName(name)
	if err != nil {
		return Project{}, err
	}

	cleanDescription := normalizeDescription(description)
	return s.repo.CreateProjectWithOwner(ctx, cleanName, cleanDescription, ownerID)
}

func (s *Service) ListProjects(ctx context.Context, userID uuid.UUID) ([]Project, error) {
	return s.repo.ListProjectsForUser(ctx, userID)
}

func (s *Service) GetProject(ctx context.Context, userID, projectID uuid.UUID) (Project, error) {
	return s.repo.GetProjectForUser(ctx, projectID, userID)
}

func (s *Service) UpdateProject(ctx context.Context, userID, projectID uuid.UUID, input UpdateProjectInput) (Project, error) {
	params, err := s.buildUpdateParams(projectID, input)
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

	return s.repo.UpdateProject(ctx, params)
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

type UpdateProjectInput struct {
	Name        *string
	Description *string
}

func (s *Service) buildUpdateParams(projectID uuid.UUID, input UpdateProjectInput) (UpdateProjectParams, error) {
	if input.Name == nil && input.Description == nil {
		return UpdateProjectParams{}, ErrInvalidProjectPatch
	}

	params := UpdateProjectParams{ProjectID: projectID}

	if input.Name != nil {
		cleanName, err := validateProjectName(*input.Name)
		if err != nil {
			return UpdateProjectParams{}, err
		}
		params.Name = &cleanName
	}

	if input.Description != nil {
		params.Description = normalizeDescription(input.Description)
	}

	return params, nil
}

func normalizeDescription(description *string) *string {
	if description == nil {
		return nil
	}
	clean := strings.TrimSpace(*description)
	return &clean
}

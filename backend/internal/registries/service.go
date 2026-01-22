package registries

import (
	"context"
	"errors"
	"net/url"
	"strings"

	"github.com/dimas1q/dockslim/backend/internal/projects"
	"github.com/google/uuid"
)

var (
	ErrInvalidRegistryName  = errors.New("invalid registry name")
	ErrInvalidRegistryURL   = errors.New("invalid registry url")
	ErrInvalidRegistryType  = errors.New("invalid registry type")
	ErrInvalidRegistryPatch = errors.New("invalid registry patch")
	ErrInvalidRegistryCreds = errors.New("invalid registry credentials")
	ErrProjectNotFound      = errors.New("project not found")
	ErrNotOwner             = errors.New("user is not project owner")
)

type RepositoryStore interface {
	ListRegistriesByProject(ctx context.Context, projectID uuid.UUID) ([]Registry, error)
	CreateRegistry(ctx context.Context, params CreateRegistryParams) (Registry, error)
	UpdateRegistry(ctx context.Context, params UpdateRegistryParams) (Registry, error)
	DeleteRegistry(ctx context.Context, projectID, registryID uuid.UUID) error
}

type MembershipStore interface {
	GetMemberRole(ctx context.Context, projectID, userID uuid.UUID) (string, error)
}

type Service struct {
	repo      RepositoryStore
	members   MembershipStore
	activeKey EncryptionKey
}

func NewService(repo RepositoryStore, members MembershipStore, activeKey EncryptionKey) *Service {
	return &Service{repo: repo, members: members, activeKey: activeKey}
}

type CreateRegistryInput struct {
	Name        string
	Type        string
	RegistryURL string
	Username    string
	Password    string
}

type UpdateRegistryInput struct {
	Name        *string
	RegistryURL *string
	Username    *string
	Token       *string
}

func (s *Service) ListRegistries(ctx context.Context, userID, projectID uuid.UUID) ([]Registry, error) {
	_, err := s.members.GetMemberRole(ctx, projectID, userID)
	if err != nil {
		if errors.Is(err, projects.ErrProjectMemberNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, err
	}

	return s.repo.ListRegistriesByProject(ctx, projectID)
}

func (s *Service) CreateRegistry(ctx context.Context, userID, projectID uuid.UUID, input CreateRegistryInput) (Registry, error) {
	role, err := s.members.GetMemberRole(ctx, projectID, userID)
	if err != nil {
		if errors.Is(err, projects.ErrProjectMemberNotFound) {
			return Registry{}, ErrProjectNotFound
		}
		return Registry{}, err
	}
	if role != projects.RoleOwner {
		return Registry{}, ErrNotOwner
	}

	params, err := s.buildCreateParams(projectID, input)
	if err != nil {
		return Registry{}, err
	}

	return s.repo.CreateRegistry(ctx, params)
}

func (s *Service) DeleteRegistry(ctx context.Context, userID, projectID, registryID uuid.UUID) error {
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

	return s.repo.DeleteRegistry(ctx, projectID, registryID)
}

func (s *Service) UpdateRegistry(ctx context.Context, userID, projectID, registryID uuid.UUID, input UpdateRegistryInput) (Registry, error) {
	role, err := s.members.GetMemberRole(ctx, projectID, userID)
	if err != nil {
		if errors.Is(err, projects.ErrProjectMemberNotFound) {
			return Registry{}, ErrProjectNotFound
		}
		return Registry{}, err
	}
	if role != projects.RoleOwner {
		return Registry{}, ErrNotOwner
	}

	params, err := s.buildUpdateParams(projectID, registryID, input)
	if err != nil {
		return Registry{}, err
	}

	return s.repo.UpdateRegistry(ctx, params)
}

func (s *Service) buildCreateParams(projectID uuid.UUID, input CreateRegistryInput) (CreateRegistryParams, error) {
	name, err := validateRegistryName(input.Name)
	if err != nil {
		return CreateRegistryParams{}, err
	}

	registryURL, err := validateRegistryURL(input.RegistryURL)
	if err != nil {
		return CreateRegistryParams{}, err
	}

	registryType, err := validateRegistryType(input.Type)
	if err != nil {
		return CreateRegistryParams{}, err
	}

	var username *string
	if strings.TrimSpace(input.Username) != "" {
		clean := strings.TrimSpace(input.Username)
		username = &clean
	}

	var passwordEnc []byte
	if strings.TrimSpace(input.Password) != "" {
		if len(s.activeKey.KeyMaterial) == 0 {
			return CreateRegistryParams{}, ErrInvalidEncryptionKey
		}
		enc, err := EncryptSecret(s.activeKey.KeyMaterial, input.Password)
		if err != nil {
			return CreateRegistryParams{}, err
		}
		passwordEnc = enc
	}

	return CreateRegistryParams{
		ProjectID:   projectID,
		Name:        name,
		Type:        registryType,
		RegistryURL: registryURL,
		Username:    username,
		PasswordEnc: passwordEnc,
	}, nil
}

func (s *Service) buildUpdateParams(projectID, registryID uuid.UUID, input UpdateRegistryInput) (UpdateRegistryParams, error) {
	if input.Name == nil && input.RegistryURL == nil && input.Username == nil && input.Token == nil {
		return UpdateRegistryParams{}, ErrInvalidRegistryPatch
	}

	params := UpdateRegistryParams{
		ProjectID:  projectID,
		RegistryID: registryID,
	}

	if input.Name != nil {
		name, err := validateRegistryName(*input.Name)
		if err != nil {
			return UpdateRegistryParams{}, err
		}
		params.Name = &name
	}

	if input.RegistryURL != nil {
		registryURL, err := validateRegistryURL(*input.RegistryURL)
		if err != nil {
			return UpdateRegistryParams{}, err
		}
		params.RegistryURL = &registryURL
	}

	if input.Username != nil || input.Token != nil {
		if input.Username == nil || input.Token == nil {
			return UpdateRegistryParams{}, ErrInvalidRegistryCreds
		}
		username := strings.TrimSpace(*input.Username)
		if username == "" {
			return UpdateRegistryParams{}, ErrInvalidRegistryCreds
		}
		token := strings.TrimSpace(*input.Token)
		if token == "" {
			return UpdateRegistryParams{}, ErrInvalidRegistryCreds
		}
		if len(s.activeKey.KeyMaterial) == 0 {
			return UpdateRegistryParams{}, ErrInvalidEncryptionKey
		}
		enc, err := EncryptSecret(s.activeKey.KeyMaterial, token)
		if err != nil {
			return UpdateRegistryParams{}, err
		}
		params.Username = &username
		params.PasswordEnc = enc
	}

	return params, nil
}

func validateRegistryName(name string) (string, error) {
	clean := strings.TrimSpace(name)
	if len(clean) < 2 || len(clean) > 100 {
		return "", ErrInvalidRegistryName
	}
	return clean, nil
}

func validateRegistryURL(raw string) (string, error) {
	clean := strings.TrimSpace(raw)
	if clean == "" {
		return "", ErrInvalidRegistryURL
	}
	parsed, err := url.ParseRequestURI(clean)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", ErrInvalidRegistryURL
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", ErrInvalidRegistryURL
	}
	return clean, nil
}

func validateRegistryType(value string) (string, error) {
	clean := strings.TrimSpace(strings.ToLower(value))
	if clean != TypeGeneric {
		return "", ErrInvalidRegistryType
	}
	return clean, nil
}

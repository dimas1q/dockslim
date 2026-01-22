package registries

import (
	"context"
	"testing"

	"github.com/dimas1q/dockslim/backend/internal/projects"
	"github.com/google/uuid"
)

type fakeRepo struct {
	registries      []Registry
	lastPasswordEnc []byte
	lastUpdateEnc   *[]byte
}

func (f *fakeRepo) ListRegistriesByProject(ctx context.Context, projectID uuid.UUID) ([]Registry, error) {
	return f.registries, nil
}

func (f *fakeRepo) CreateRegistry(ctx context.Context, params CreateRegistryParams) (Registry, error) {
	registry := Registry{
		ID:          uuid.New(),
		ProjectID:   params.ProjectID,
		Name:        params.Name,
		Type:        params.Type,
		RegistryURL: params.RegistryURL,
		Username:    params.Username,
	}
	f.registries = append(f.registries, registry)
	f.lastPasswordEnc = params.PasswordEnc
	return registry, nil
}

func (f *fakeRepo) DeleteRegistry(ctx context.Context, projectID, registryID uuid.UUID) error {
	return nil
}

func (f *fakeRepo) UpdateRegistry(ctx context.Context, params UpdateRegistryParams) (Registry, error) {
	f.lastUpdateEnc = params.PasswordEnc
	for i, registry := range f.registries {
		if registry.ID != params.RegistryID || registry.ProjectID != params.ProjectID {
			continue
		}
		if params.Name != nil {
			registry.Name = *params.Name
		}
		if params.RegistryURL != nil {
			registry.RegistryURL = *params.RegistryURL
		}
		if params.Username != nil {
			registry.Username = params.Username
		}
		f.registries[i] = registry
		return registry, nil
	}
	return Registry{}, ErrRegistryNotFound
}

type fakeMembership struct {
	role string
	err  error
}

func (f *fakeMembership) GetMemberRole(ctx context.Context, projectID, userID uuid.UUID) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	return f.role, nil
}

func TestServiceCreateRegistryOwnerOnly(t *testing.T) {
	repo := &fakeRepo{}
	members := &fakeMembership{role: "member"}
	service := NewService(repo, members, EncryptionKey{KeyMaterial: []byte("01234567890123456789012345678901")})

	_, err := service.CreateRegistry(context.Background(), uuid.New(), uuid.New(), CreateRegistryInput{
		Name:        "My Registry",
		Type:        "generic",
		RegistryURL: "https://registry.example.com",
	})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if err != ErrNotOwner {
		t.Fatalf("expected ErrNotOwner, got %v", err)
	}
}

func TestServiceListRegistriesMemberAllowed(t *testing.T) {
	repo := &fakeRepo{}
	members := &fakeMembership{role: projects.RoleOwner}
	service := NewService(repo, members, EncryptionKey{KeyMaterial: []byte("01234567890123456789012345678901")})

	projectID := uuid.New()
	repo.registries = []Registry{{ID: uuid.New(), ProjectID: projectID, Name: "One", Type: TypeGeneric, RegistryURL: "https://example.com"}}

	items, err := service.ListRegistries(context.Background(), uuid.New(), projectID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 registry, got %d", len(items))
	}
}

func TestServiceListRegistriesNonMemberNotFound(t *testing.T) {
	repo := &fakeRepo{}
	members := &fakeMembership{err: projects.ErrProjectMemberNotFound}
	service := NewService(repo, members, EncryptionKey{KeyMaterial: []byte("01234567890123456789012345678901")})

	_, err := service.ListRegistries(context.Background(), uuid.New(), uuid.New())
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if err != ErrProjectNotFound {
		t.Fatalf("expected ErrProjectNotFound, got %v", err)
	}
}

func TestServiceUpdateRegistryDoesNotTouchPasswordOnMetadataChange(t *testing.T) {
	projectID := uuid.New()
	registryID := uuid.New()
	repo := &fakeRepo{
		registries: []Registry{
			{
				ID:          registryID,
				ProjectID:   projectID,
				Name:        "Registry",
				Type:        TypeGeneric,
				RegistryURL: "https://registry.example.com",
			},
		},
	}
	members := &fakeMembership{role: projects.RoleOwner}
	service := NewService(repo, members, EncryptionKey{KeyMaterial: []byte("01234567890123456789012345678901")})

	_, err := service.UpdateRegistry(context.Background(), uuid.New(), projectID, registryID, UpdateRegistryInput{
		Name:        ptr("Updated Registry"),
		RegistryURL: ptr("https://new.registry.example.com"),
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.lastUpdateEnc != nil {
		t.Fatalf("expected password not to be updated")
	}
}

func ptr(value string) *string {
	return &value
}

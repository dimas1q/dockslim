package projects

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

type fakeRepo struct {
	projects map[uuid.UUID]Project
	members  map[uuid.UUID]map[uuid.UUID]string
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		projects: make(map[uuid.UUID]Project),
		members:  make(map[uuid.UUID]map[uuid.UUID]string),
	}
}

func (f *fakeRepo) CreateProjectWithOwner(ctx context.Context, name string, ownerID uuid.UUID) (Project, error) {
	project := Project{
		ID:   uuid.New(),
		Name: name,
	}
	f.projects[project.ID] = project
	if f.members[project.ID] == nil {
		f.members[project.ID] = make(map[uuid.UUID]string)
	}
	f.members[project.ID][ownerID] = RoleOwner
	return project, nil
}

func (f *fakeRepo) ListProjectsForUser(ctx context.Context, userID uuid.UUID) ([]Project, error) {
	var results []Project
	for projectID, members := range f.members {
		if _, ok := members[userID]; ok {
			results = append(results, f.projects[projectID])
		}
	}
	return results, nil
}

func (f *fakeRepo) GetProjectForUser(ctx context.Context, projectID, userID uuid.UUID) (Project, error) {
	members, ok := f.members[projectID]
	if !ok {
		return Project{}, ErrProjectNotFound
	}
	if _, ok := members[userID]; !ok {
		return Project{}, ErrProjectNotFound
	}
	project, ok := f.projects[projectID]
	if !ok {
		return Project{}, ErrProjectNotFound
	}
	return project, nil
}

func (f *fakeRepo) GetMemberRole(ctx context.Context, projectID, userID uuid.UUID) (string, error) {
	members, ok := f.members[projectID]
	if !ok {
		return "", ErrProjectMemberNotFound
	}
	role, ok := members[userID]
	if !ok {
		return "", ErrProjectMemberNotFound
	}
	return role, nil
}

func (f *fakeRepo) UpdateProjectName(ctx context.Context, projectID uuid.UUID, name string) (Project, error) {
	project, ok := f.projects[projectID]
	if !ok {
		return Project{}, ErrProjectNotFound
	}
	project.Name = name
	f.projects[projectID] = project
	return project, nil
}

func (f *fakeRepo) DeleteProject(ctx context.Context, projectID uuid.UUID) error {
	if _, ok := f.projects[projectID]; !ok {
		return ErrProjectNotFound
	}
	delete(f.projects, projectID)
	delete(f.members, projectID)
	return nil
}

func TestServiceCreateProject(t *testing.T) {
	repo := newFakeRepo()
	service := NewService(repo)

	ownerID := uuid.New()
	project, err := service.CreateProject(context.Background(), ownerID, "  My Project  ")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if project.Name != "My Project" {
		t.Fatalf("expected project name to be trimmed, got %q", project.Name)
	}

	role, err := repo.GetMemberRole(context.Background(), project.ID, ownerID)
	if err != nil {
		t.Fatalf("expected owner membership to be created, got %v", err)
	}
	if role != RoleOwner {
		t.Fatalf("expected owner role, got %q", role)
	}
}

func TestServiceListProjectsFiltersByMember(t *testing.T) {
	repo := newFakeRepo()
	service := NewService(repo)

	ownerID := uuid.New()
	otherID := uuid.New()

	projectA, _ := service.CreateProject(context.Background(), ownerID, "Project A")
	projectB, _ := service.CreateProject(context.Background(), otherID, "Project B")

	projects, err := service.ListProjects(context.Background(), ownerID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}
	if projects[0].ID != projectA.ID {
		t.Fatalf("expected project %s, got %s", projectA.ID, projects[0].ID)
	}
	if projects[0].ID == projectB.ID {
		t.Fatalf("did not expect projectB in results")
	}
}

func TestServiceGetProjectAsNonMemberReturnsNotFound(t *testing.T) {
	repo := newFakeRepo()
	service := NewService(repo)

	ownerID := uuid.New()
	project, _ := service.CreateProject(context.Background(), ownerID, "Private Project")

	nonMemberID := uuid.New()
	_, err := service.GetProject(context.Background(), nonMemberID, project.ID)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if err != ErrProjectNotFound {
		t.Fatalf("expected ErrProjectNotFound, got %v", err)
	}
}

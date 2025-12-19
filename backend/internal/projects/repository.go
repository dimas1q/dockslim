package projects

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
)

var (
	ErrProjectNotFound       = errors.New("project not found")
	ErrProjectMemberNotFound = errors.New("project member not found")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateProjectWithOwner(ctx context.Context, name string, ownerID uuid.UUID) (Project, error) {
	const projectQuery = `
		INSERT INTO projects (name)
		VALUES ($1)
		RETURNING id, name, created_at, updated_at
	`
	const memberQuery = `
		INSERT INTO project_members (project_id, user_id, role)
		VALUES ($1, $2, $3)
	`

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Project{}, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var project Project
	err = tx.QueryRowContext(ctx, projectQuery, name).Scan(
		&project.ID,
		&project.Name,
		&project.CreatedAt,
		&project.UpdatedAt,
	)
	if err != nil {
		return Project{}, err
	}

	_, err = tx.ExecContext(ctx, memberQuery, project.ID, ownerID, RoleOwner)
	if err != nil {
		return Project{}, err
	}

	if err = tx.Commit(); err != nil {
		return Project{}, err
	}

	return project, nil
}

func (r *Repository) ListProjectsForUser(ctx context.Context, userID uuid.UUID) ([]Project, error) {
	const query = `
		SELECT p.id, p.name, p.created_at, p.updated_at
		FROM projects p
		JOIN project_members pm ON pm.project_id = p.id
		WHERE pm.user_id = $1
		ORDER BY p.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var project Project
		if err := rows.Scan(&project.ID, &project.Name, &project.CreatedAt, &project.UpdatedAt); err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return projects, nil
}

func (r *Repository) GetProjectForUser(ctx context.Context, projectID, userID uuid.UUID) (Project, error) {
	const query = `
		SELECT p.id, p.name, p.created_at, p.updated_at
		FROM projects p
		JOIN project_members pm ON pm.project_id = p.id
		WHERE p.id = $1 AND pm.user_id = $2
	`

	var project Project
	err := r.db.QueryRowContext(ctx, query, projectID, userID).Scan(
		&project.ID,
		&project.Name,
		&project.CreatedAt,
		&project.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Project{}, ErrProjectNotFound
	}
	if err != nil {
		return Project{}, err
	}

	return project, nil
}

func (r *Repository) GetMemberRole(ctx context.Context, projectID, userID uuid.UUID) (string, error) {
	const query = `
		SELECT role
		FROM project_members
		WHERE project_id = $1 AND user_id = $2
	`

	var role string
	err := r.db.QueryRowContext(ctx, query, projectID, userID).Scan(&role)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrProjectMemberNotFound
	}
	if err != nil {
		return "", err
	}

	return role, nil
}

func (r *Repository) UpdateProjectName(ctx context.Context, projectID uuid.UUID, name string) (Project, error) {
	const query = `
		UPDATE projects
		SET name = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING id, name, created_at, updated_at
	`

	var project Project
	err := r.db.QueryRowContext(ctx, query, projectID, name).Scan(
		&project.ID,
		&project.Name,
		&project.CreatedAt,
		&project.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Project{}, ErrProjectNotFound
	}
	if err != nil {
		return Project{}, err
	}

	return project, nil
}

func (r *Repository) DeleteProject(ctx context.Context, projectID uuid.UUID) error {
	const query = `
		DELETE FROM projects
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, projectID)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrProjectNotFound
	}

	return nil
}

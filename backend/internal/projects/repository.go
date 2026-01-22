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

type UpdateProjectParams struct {
	ProjectID   uuid.UUID
	Name        *string
	Description *string
}

func (r *Repository) CreateProjectWithOwner(ctx context.Context, name string, description *string, ownerID uuid.UUID) (Project, error) {
	const projectQuery = `
		INSERT INTO projects (name, description)
		VALUES ($1, $2)
		RETURNING id, name, description, created_at, updated_at
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
	var descriptionValue sql.NullString
	if description != nil {
		descriptionValue = sql.NullString{String: *description, Valid: true}
	}
	err = tx.QueryRowContext(ctx, projectQuery, name, descriptionValue).Scan(
		&project.ID,
		&project.Name,
		&descriptionValue,
		&project.CreatedAt,
		&project.UpdatedAt,
	)
	if err != nil {
		return Project{}, err
	}
	if descriptionValue.Valid {
		project.Description = &descriptionValue.String
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
		SELECT p.id, p.name, p.description, p.created_at, p.updated_at
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
		var description sql.NullString
		if err := rows.Scan(&project.ID, &project.Name, &description, &project.CreatedAt, &project.UpdatedAt); err != nil {
			return nil, err
		}
		if description.Valid {
			project.Description = &description.String
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
		SELECT p.id, p.name, p.description, p.created_at, p.updated_at
		FROM projects p
		JOIN project_members pm ON pm.project_id = p.id
		WHERE p.id = $1 AND pm.user_id = $2
	`

	var project Project
	var description sql.NullString
	err := r.db.QueryRowContext(ctx, query, projectID, userID).Scan(
		&project.ID,
		&project.Name,
		&description,
		&project.CreatedAt,
		&project.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Project{}, ErrProjectNotFound
	}
	if err != nil {
		return Project{}, err
	}

	if description.Valid {
		project.Description = &description.String
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

func (r *Repository) UpdateProject(ctx context.Context, params UpdateProjectParams) (Project, error) {
	const query = `
		UPDATE projects
		SET name = COALESCE($2, name),
			description = COALESCE($3, description),
			updated_at = NOW()
		WHERE id = $1
		RETURNING id, name, description, created_at, updated_at
	`

	var project Project
	var name sql.NullString
	if params.Name != nil {
		name = sql.NullString{String: *params.Name, Valid: true}
	}
	var description sql.NullString
	if params.Description != nil {
		description = sql.NullString{String: *params.Description, Valid: true}
	}
	err := r.db.QueryRowContext(ctx, query, params.ProjectID, name, description).Scan(
		&project.ID,
		&project.Name,
		&description,
		&project.CreatedAt,
		&project.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Project{}, ErrProjectNotFound
	}
	if err != nil {
		return Project{}, err
	}

	if description.Valid {
		project.Description = &description.String
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

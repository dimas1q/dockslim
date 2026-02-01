package registries

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
)

var (
	ErrKeyNotFound          = errors.New("encryption key not found")
	ErrRegistryNotFound     = errors.New("registry not found")
	ErrRegistryNameConflict = errors.New("registry name already exists")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

type CreateRegistryParams struct {
	ProjectID   uuid.UUID
	Name        string
	Type        string
	RegistryURL string
	Username    *string
	PasswordEnc []byte
}

type UpdateRegistryParams struct {
	ProjectID   uuid.UUID
	RegistryID  uuid.UUID
	Name        *string
	RegistryURL *string
	Username    *string
	PasswordEnc *[]byte
}

func (r *Repository) GetRegistryByName(ctx context.Context, projectID uuid.UUID, name string) (Registry, error) {
	const query = `
		SELECT id, project_id, name, type, registry_url, username, password_enc, created_at, updated_at
		FROM registries
		WHERE project_id = $1 AND name = $2
	`

	var registry Registry
	var username sql.NullString
	var passwordEnc []byte
	err := r.db.QueryRowContext(ctx, query, projectID, name).Scan(
		&registry.ID,
		&registry.ProjectID,
		&registry.Name,
		&registry.Type,
		&registry.RegistryURL,
		&username,
		&passwordEnc,
		&registry.CreatedAt,
		&registry.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Registry{}, ErrRegistryNotFound
	}
	if err != nil {
		return Registry{}, err
	}
	if username.Valid {
		registry.Username = &username.String
	}
	return registry, nil
}

func (r *Repository) GetActiveKey(ctx context.Context) (EncryptionKey, error) {
	const query = `
		SELECT id, key_id, key_material, is_active, created_at
		FROM encryption_keys
		WHERE is_active = TRUE
		ORDER BY created_at DESC
		LIMIT 1
	`

	var key EncryptionKey
	err := r.db.QueryRowContext(ctx, query).Scan(
		&key.ID,
		&key.KeyID,
		&key.KeyMaterial,
		&key.IsActive,
		&key.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return EncryptionKey{}, ErrKeyNotFound
	}
	if err != nil {
		return EncryptionKey{}, err
	}

	return key, nil
}

func (r *Repository) CreateKey(ctx context.Context, key EncryptionKey) (EncryptionKey, error) {
	const deactivateQuery = `
		UPDATE encryption_keys
		SET is_active = FALSE
		WHERE is_active = TRUE
	`
	const insertQuery = `
		INSERT INTO encryption_keys (key_id, key_material, is_active)
		VALUES ($1, $2, $3)
		RETURNING id, key_id, key_material, is_active, created_at
	`

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return EncryptionKey{}, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.ExecContext(ctx, deactivateQuery); err != nil {
		return EncryptionKey{}, err
	}

	var created EncryptionKey
	err = tx.QueryRowContext(ctx, insertQuery, key.KeyID, key.KeyMaterial, key.IsActive).Scan(
		&created.ID,
		&created.KeyID,
		&created.KeyMaterial,
		&created.IsActive,
		&created.CreatedAt,
	)
	if err != nil {
		return EncryptionKey{}, err
	}

	if err = tx.Commit(); err != nil {
		return EncryptionKey{}, err
	}

	return created, nil
}

func (r *Repository) CreateRegistry(ctx context.Context, params CreateRegistryParams) (Registry, error) {
	const query = `
		INSERT INTO registries (project_id, name, type, registry_url, username, password_enc)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, project_id, name, type, registry_url, username, created_at, updated_at
	`

	var username sql.NullString
	if params.Username != nil && *params.Username != "" {
		username = sql.NullString{String: *params.Username, Valid: true}
	}

	var registry Registry
	var usernameOut sql.NullString
	err := r.db.QueryRowContext(
		ctx,
		query,
		params.ProjectID,
		params.Name,
		params.Type,
		params.RegistryURL,
		username,
		params.PasswordEnc,
	).Scan(
		&registry.ID,
		&registry.ProjectID,
		&registry.Name,
		&registry.Type,
		&registry.RegistryURL,
		&usernameOut,
		&registry.CreatedAt,
		&registry.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return Registry{}, ErrRegistryNameConflict
		}
		return Registry{}, err
	}

	if usernameOut.Valid {
		registry.Username = &usernameOut.String
	}

	return registry, nil
}

func (r *Repository) ListRegistriesByProject(ctx context.Context, projectID uuid.UUID) ([]Registry, error) {
	const query = `
		SELECT id, project_id, name, type, registry_url, username, created_at, updated_at
		FROM registries
		WHERE project_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var registries []Registry
	for rows.Next() {
		var registry Registry
		var username sql.NullString
		if err := rows.Scan(
			&registry.ID,
			&registry.ProjectID,
			&registry.Name,
			&registry.Type,
			&registry.RegistryURL,
			&username,
			&registry.CreatedAt,
			&registry.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if username.Valid {
			registry.Username = &username.String
		}
		registries = append(registries, registry)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return registries, nil
}

func (r *Repository) GetRegistryForProject(ctx context.Context, projectID, registryID uuid.UUID) (Registry, error) {
	const query = `
		SELECT id, project_id, name, type, registry_url, username, created_at, updated_at
		FROM registries
		WHERE id = $1 AND project_id = $2
	`

	var registry Registry
	var username sql.NullString
	err := r.db.QueryRowContext(ctx, query, registryID, projectID).Scan(
		&registry.ID,
		&registry.ProjectID,
		&registry.Name,
		&registry.Type,
		&registry.RegistryURL,
		&username,
		&registry.CreatedAt,
		&registry.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Registry{}, ErrRegistryNotFound
	}
	if err != nil {
		return Registry{}, err
	}

	if username.Valid {
		registry.Username = &username.String
	}

	return registry, nil
}

func (r *Repository) DeleteRegistry(ctx context.Context, projectID, registryID uuid.UUID) error {
	const query = `
		DELETE FROM registries
		WHERE id = $1 AND project_id = $2
	`

	result, err := r.db.ExecContext(ctx, query, registryID, projectID)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrRegistryNotFound
	}

	return nil
}

func (r *Repository) UpdateRegistry(ctx context.Context, params UpdateRegistryParams) (Registry, error) {
	const query = `
		UPDATE registries
		SET name = COALESCE($1, name),
			registry_url = COALESCE($2, registry_url),
			username = COALESCE($3, username),
			password_enc = COALESCE($4, password_enc),
			updated_at = NOW()
		WHERE id = $5 AND project_id = $6
		RETURNING id, project_id, name, type, registry_url, username, created_at, updated_at
	`

	var name sql.NullString
	if params.Name != nil && *params.Name != "" {
		name = sql.NullString{String: *params.Name, Valid: true}
	}
	var registryURL sql.NullString
	if params.RegistryURL != nil && *params.RegistryURL != "" {
		registryURL = sql.NullString{String: *params.RegistryURL, Valid: true}
	}
	var username sql.NullString
	if params.Username != nil && *params.Username != "" {
		username = sql.NullString{String: *params.Username, Valid: true}
	}
	var passwordEnc interface{}
	if params.PasswordEnc != nil {
		passwordEnc = *params.PasswordEnc
	}

	var registry Registry
	var usernameOut sql.NullString
	err := r.db.QueryRowContext(
		ctx,
		query,
		name,
		registryURL,
		username,
		passwordEnc,
		params.RegistryID,
		params.ProjectID,
	).Scan(
		&registry.ID,
		&registry.ProjectID,
		&registry.Name,
		&registry.Type,
		&registry.RegistryURL,
		&usernameOut,
		&registry.CreatedAt,
		&registry.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Registry{}, ErrRegistryNotFound
	}
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return Registry{}, ErrRegistryNameConflict
		}
		return Registry{}, err
	}

	if usernameOut.Valid {
		registry.Username = &usernameOut.String
	}

	return registry, nil
}

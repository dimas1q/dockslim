package budgets

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
)

var (
	ErrBudgetNotFound = errors.New("budget not found")
	ErrBudgetConflict = errors.New("budget already exists")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// ListBudgetsByProject returns default (image null) and overrides (image not null) ordered by image.
func (r *Repository) ListBudgetsByProject(ctx context.Context, projectID uuid.UUID) ([]Budget, error) {
	const query = `
        SELECT id, project_id, image, warn_delta_bytes, fail_delta_bytes, hard_limit_bytes, created_at, updated_at
        FROM project_budgets
        WHERE project_id = $1
        ORDER BY (image IS NULL) DESC, image ASC NULLS LAST
    `
	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var budgets []Budget
	for rows.Next() {
		var b Budget
		var image sql.NullString
		var warn sql.NullInt64
		var fail sql.NullInt64
		var limit sql.NullInt64
		if err := rows.Scan(&b.ID, &b.ProjectID, &image, &warn, &fail, &limit, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, err
		}
		if image.Valid {
			b.Image = &image.String
		}
		if warn.Valid {
			v := warn.Int64
			b.WarnDeltaBytes = &v
		}
		if fail.Valid {
			v := fail.Int64
			b.FailDeltaBytes = &v
		}
		if limit.Valid {
			v := limit.Int64
			b.HardLimitBytes = &v
		}
		budgets = append(budgets, b)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return budgets, nil
}

// UpsertDefaultBudget inserts or updates the project default (image NULL) budget.
func (r *Repository) UpsertDefaultBudget(ctx context.Context, projectID uuid.UUID, thresholds ResolvedBudget) (Budget, error) {
	const query = `
		INSERT INTO project_budgets (id, project_id, image, warn_delta_bytes, fail_delta_bytes, hard_limit_bytes)
		VALUES ($1, $2, NULL, $3, $4, $5)
		ON CONFLICT (project_id) WHERE image IS NULL
		DO UPDATE SET warn_delta_bytes = EXCLUDED.warn_delta_bytes,
		              fail_delta_bytes = EXCLUDED.fail_delta_bytes,
		              hard_limit_bytes = EXCLUDED.hard_limit_bytes,
		              updated_at = NOW()
		RETURNING id, project_id, image, warn_delta_bytes, fail_delta_bytes, hard_limit_bytes, created_at, updated_at
	`

	id := uuid.New()
	var b Budget
	var image sql.NullString
	var warn, fail, limit sql.NullInt64
	err := r.db.QueryRowContext(ctx, query, id, projectID, thresholds.WarnDeltaBytes, thresholds.FailDeltaBytes, thresholds.HardLimitBytes).
		Scan(&b.ID, &b.ProjectID, &image, &warn, &fail, &limit, &b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		return Budget{}, err
	}
	if image.Valid {
		b.Image = &image.String
	}
	if warn.Valid {
		v := warn.Int64
		b.WarnDeltaBytes = &v
	}
	if fail.Valid {
		v := fail.Int64
		b.FailDeltaBytes = &v
	}
	if limit.Valid {
		v := limit.Int64
		b.HardLimitBytes = &v
	}
	return b, nil
}

// CreateBudgetOverride inserts a per-image override.
func (r *Repository) CreateBudgetOverride(ctx context.Context, projectID uuid.UUID, image string, thresholds ResolvedBudget) (Budget, error) {
	const query = `
        INSERT INTO project_budgets (id, project_id, image, warn_delta_bytes, fail_delta_bytes, hard_limit_bytes)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id, project_id, image, warn_delta_bytes, fail_delta_bytes, hard_limit_bytes, created_at, updated_at
    `

	id := uuid.New()
	var b Budget
	var imageOut sql.NullString
	var warn, fail, limit sql.NullInt64
	err := r.db.QueryRowContext(ctx, query, id, projectID, image, thresholds.WarnDeltaBytes, thresholds.FailDeltaBytes, thresholds.HardLimitBytes).
		Scan(&b.ID, &b.ProjectID, &imageOut, &warn, &fail, &limit, &b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return Budget{}, ErrBudgetConflict
		}
		return Budget{}, err
	}
	if imageOut.Valid {
		b.Image = &imageOut.String
	}
	if warn.Valid {
		v := warn.Int64
		b.WarnDeltaBytes = &v
	}
	if fail.Valid {
		v := fail.Int64
		b.FailDeltaBytes = &v
	}
	if limit.Valid {
		v := limit.Int64
		b.HardLimitBytes = &v
	}
	return b, nil
}

// UpdateBudget updates thresholds and optionally image for a specific budget row.
func (r *Repository) UpdateBudget(ctx context.Context, budgetID, projectID uuid.UUID, image *string, thresholds ResolvedBudget) (Budget, error) {
	const query = `
        UPDATE project_budgets
        SET image = COALESCE($3, image),
            warn_delta_bytes = $4,
            fail_delta_bytes = $5,
            hard_limit_bytes = $6,
            updated_at = NOW()
        WHERE id = $1 AND project_id = $2
        RETURNING id, project_id, image, warn_delta_bytes, fail_delta_bytes, hard_limit_bytes, created_at, updated_at
    `

	var b Budget
	var imageOut sql.NullString
	var warn, fail, limit sql.NullInt64
	err := r.db.QueryRowContext(ctx, query, budgetID, projectID, image, thresholds.WarnDeltaBytes, thresholds.FailDeltaBytes, thresholds.HardLimitBytes).
		Scan(&b.ID, &b.ProjectID, &imageOut, &warn, &fail, &limit, &b.CreatedAt, &b.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Budget{}, ErrBudgetNotFound
	}
	if err != nil {
		return Budget{}, err
	}
	if imageOut.Valid {
		b.Image = &imageOut.String
	}
	if warn.Valid {
		v := warn.Int64
		b.WarnDeltaBytes = &v
	}
	if fail.Valid {
		v := fail.Int64
		b.FailDeltaBytes = &v
	}
	if limit.Valid {
		v := limit.Int64
		b.HardLimitBytes = &v
	}
	return b, nil
}

// DeleteBudget removes a budget row scoped to project.
func (r *Repository) DeleteBudget(ctx context.Context, budgetID, projectID uuid.UUID) error {
	const query = `
        DELETE FROM project_budgets
        WHERE id = $1 AND project_id = $2
    `

	result, err := r.db.ExecContext(ctx, query, budgetID, projectID)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrBudgetNotFound
	}
	return nil
}

// ResolveBudgetForImage returns an override for the image if present, otherwise the default. Returns nil if none.
func (r *Repository) ResolveBudgetForImage(ctx context.Context, projectID uuid.UUID, image string) (*ResolvedBudget, error) {
	const query = `
        SELECT warn_delta_bytes, fail_delta_bytes, hard_limit_bytes
        FROM project_budgets
        WHERE project_id = $1 AND (image = $2 OR image IS NULL)
        ORDER BY CASE WHEN image = $2 THEN 0 ELSE 1 END
        LIMIT 1
    `
	var warn, fail, limit sql.NullInt64
	err := r.db.QueryRowContext(ctx, query, projectID, image).Scan(&warn, &fail, &limit)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	result := ResolvedBudget{}
	if warn.Valid {
		v := warn.Int64
		result.WarnDeltaBytes = &v
	}
	if fail.Valid {
		v := fail.Int64
		result.FailDeltaBytes = &v
	}
	if limit.Valid {
		v := limit.Int64
		result.HardLimitBytes = &v
	}
	return &result, nil
}

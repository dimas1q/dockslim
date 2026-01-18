package analyses

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
)

var ErrAnalysisNotFound = errors.New("analysis not found")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

type CreateAnalysisParams struct {
	ProjectID      uuid.UUID
	RegistryID     *uuid.UUID
	Image          string
	Tag            string
	Status         string
	TotalSizeBytes *int64
}

func (r *Repository) CreateAnalysis(ctx context.Context, params CreateAnalysisParams) (ImageAnalysis, error) {
	const query = `
		INSERT INTO image_analyses (project_id, registry_id, image, tag, status, total_size_bytes)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, project_id, registry_id, image, tag, status, total_size_bytes, created_at, updated_at
	`

	var registryID sql.NullString
	if params.RegistryID != nil {
		registryID = sql.NullString{String: params.RegistryID.String(), Valid: true}
	}

	var totalSize sql.NullInt64
	if params.TotalSizeBytes != nil {
		totalSize = sql.NullInt64{Int64: *params.TotalSizeBytes, Valid: true}
	}

	var analysis ImageAnalysis
	var registryIDOut sql.NullString
	var totalSizeOut sql.NullInt64
	err := r.db.QueryRowContext(
		ctx,
		query,
		params.ProjectID,
		registryID,
		params.Image,
		params.Tag,
		params.Status,
		totalSize,
	).Scan(
		&analysis.ID,
		&analysis.ProjectID,
		&registryIDOut,
		&analysis.Image,
		&analysis.Tag,
		&analysis.Status,
		&totalSizeOut,
		&analysis.CreatedAt,
		&analysis.UpdatedAt,
	)
	if err != nil {
		return ImageAnalysis{}, err
	}

	if registryIDOut.Valid {
		parsed, err := uuid.Parse(registryIDOut.String)
		if err != nil {
			return ImageAnalysis{}, err
		}
		analysis.RegistryID = &parsed
	}
	if totalSizeOut.Valid {
		value := totalSizeOut.Int64
		analysis.TotalSizeBytes = &value
	}

	return analysis, nil
}

func (r *Repository) ListAnalysesByProject(ctx context.Context, projectID uuid.UUID) ([]ImageAnalysis, error) {
	const query = `
		SELECT id, project_id, registry_id, image, tag, status, total_size_bytes, created_at, updated_at
		FROM image_analyses
		WHERE project_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var analyses []ImageAnalysis
	for rows.Next() {
		var analysis ImageAnalysis
		var registryID sql.NullString
		var totalSize sql.NullInt64
		if err := rows.Scan(
			&analysis.ID,
			&analysis.ProjectID,
			&registryID,
			&analysis.Image,
			&analysis.Tag,
			&analysis.Status,
			&totalSize,
			&analysis.CreatedAt,
			&analysis.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if registryID.Valid {
			parsed, err := uuid.Parse(registryID.String)
			if err != nil {
				return nil, err
			}
			analysis.RegistryID = &parsed
		}
		if totalSize.Valid {
			value := totalSize.Int64
			analysis.TotalSizeBytes = &value
		}
		analyses = append(analyses, analysis)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return analyses, nil
}

func (r *Repository) GetAnalysisForProject(ctx context.Context, projectID, analysisID uuid.UUID) (ImageAnalysis, error) {
	const query = `
		SELECT id, project_id, registry_id, image, tag, status, total_size_bytes, created_at, updated_at
		FROM image_analyses
		WHERE id = $1 AND project_id = $2
	`

	var analysis ImageAnalysis
	var registryID sql.NullString
	var totalSize sql.NullInt64
	err := r.db.QueryRowContext(ctx, query, analysisID, projectID).Scan(
		&analysis.ID,
		&analysis.ProjectID,
		&registryID,
		&analysis.Image,
		&analysis.Tag,
		&analysis.Status,
		&totalSize,
		&analysis.CreatedAt,
		&analysis.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return ImageAnalysis{}, ErrAnalysisNotFound
	}
	if err != nil {
		return ImageAnalysis{}, err
	}

	if registryID.Valid {
		parsed, err := uuid.Parse(registryID.String)
		if err != nil {
			return ImageAnalysis{}, err
		}
		analysis.RegistryID = &parsed
	}
	if totalSize.Valid {
		value := totalSize.Int64
		analysis.TotalSizeBytes = &value
	}

	return analysis, nil
}

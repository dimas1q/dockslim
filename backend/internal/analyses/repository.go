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
	GitRef         *string
	CommitSHA      *string
	Status         string
	TotalSizeBytes *int64
}

func (r *Repository) CreateAnalysis(ctx context.Context, params CreateAnalysisParams) (ImageAnalysis, error) {
	const query = `
		INSERT INTO image_analyses (project_id, registry_id, image, tag, git_ref, commit_sha, status, total_size_bytes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, project_id, registry_id, image, tag, git_ref, commit_sha, status, total_size_bytes, layer_count, largest_layer_bytes, result_json, started_at, finished_at, analyzed_at, created_at, updated_at
	`
	const jobQuery = `
		INSERT INTO analysis_jobs (analysis_id, status)
		VALUES ($1, $2)
	`

	var registryID interface{}
	if params.RegistryID != nil {
		registryID = *params.RegistryID
	}

	var totalSize sql.NullInt64
	if params.TotalSizeBytes != nil {
		totalSize = sql.NullInt64{Int64: *params.TotalSizeBytes, Valid: true}
	}

	var gitRef sql.NullString
	if params.GitRef != nil && *params.GitRef != "" {
		gitRef = sql.NullString{String: *params.GitRef, Valid: true}
	}
	var commitSHA sql.NullString
	if params.CommitSHA != nil && *params.CommitSHA != "" {
		commitSHA = sql.NullString{String: *params.CommitSHA, Valid: true}
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return ImageAnalysis{}, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var analysis ImageAnalysis
	var registryIDOut sql.NullString
	var totalSizeOut sql.NullInt64
	var layerCountOut sql.NullInt64
	var largestLayerOut sql.NullInt64
	var resultJSON sql.NullString
	var startedAt sql.NullTime
	var finishedAt sql.NullTime
	var analyzedAt sql.NullTime
	var gitRefOut sql.NullString
	var commitSHAOut sql.NullString
	err = tx.QueryRowContext(
		ctx,
		query,
		params.ProjectID,
		registryID,
		params.Image,
		params.Tag,
		gitRef,
		commitSHA,
		params.Status,
		totalSize,
	).Scan(
		&analysis.ID,
		&analysis.ProjectID,
		&registryIDOut,
		&analysis.Image,
		&analysis.Tag,
		&gitRefOut,
		&commitSHAOut,
		&analysis.Status,
		&totalSizeOut,
		&layerCountOut,
		&largestLayerOut,
		&resultJSON,
		&startedAt,
		&finishedAt,
		&analyzedAt,
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
	if layerCountOut.Valid {
		value := int(layerCountOut.Int64)
		analysis.LayerCount = &value
	}
	if largestLayerOut.Valid {
		value := largestLayerOut.Int64
		analysis.LargestLayerBytes = &value
	}
	if resultJSON.Valid {
		analysis.ResultJSON = []byte(resultJSON.String)
	}
	if startedAt.Valid {
		value := startedAt.Time
		analysis.StartedAt = &value
	}
	if finishedAt.Valid {
		value := finishedAt.Time
		analysis.FinishedAt = &value
	}
	if analyzedAt.Valid {
		value := analyzedAt.Time
		analysis.AnalyzedAt = &value
	}
	if gitRefOut.Valid {
		value := gitRefOut.String
		analysis.GitRef = &value
	}
	if commitSHAOut.Valid {
		value := commitSHAOut.String
		analysis.CommitSHA = &value
	}

	if _, err = tx.ExecContext(ctx, jobQuery, analysis.ID, params.Status); err != nil {
		return ImageAnalysis{}, err
	}

	if err = tx.Commit(); err != nil {
		return ImageAnalysis{}, err
	}

	return analysis, nil
}

func (r *Repository) ListAnalysesByProject(ctx context.Context, projectID uuid.UUID) ([]ImageAnalysis, error) {
	const query = `
		SELECT id, project_id, registry_id, image, tag, git_ref, commit_sha, status, total_size_bytes, layer_count, largest_layer_bytes, result_json, started_at, finished_at, analyzed_at, created_at, updated_at
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
		var layerCount sql.NullInt64
		var largestLayer sql.NullInt64
		var gitRef sql.NullString
		var commitSHA sql.NullString
		var resultJSON sql.NullString
		var startedAt sql.NullTime
		var finishedAt sql.NullTime
		var analyzedAt sql.NullTime
		if err := rows.Scan(
			&analysis.ID,
			&analysis.ProjectID,
			&registryID,
			&analysis.Image,
			&analysis.Tag,
			&gitRef,
			&commitSHA,
			&analysis.Status,
			&totalSize,
			&layerCount,
			&largestLayer,
			&resultJSON,
			&startedAt,
			&finishedAt,
			&analyzedAt,
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
		if layerCount.Valid {
			value := int(layerCount.Int64)
			analysis.LayerCount = &value
		}
		if largestLayer.Valid {
			value := largestLayer.Int64
			analysis.LargestLayerBytes = &value
		}
		if resultJSON.Valid {
			analysis.ResultJSON = []byte(resultJSON.String)
		}
		if startedAt.Valid {
			value := startedAt.Time
			analysis.StartedAt = &value
		}
		if finishedAt.Valid {
			value := finishedAt.Time
			analysis.FinishedAt = &value
		}
		if analyzedAt.Valid {
			value := analyzedAt.Time
			analysis.AnalyzedAt = &value
		}
		if gitRef.Valid {
			value := gitRef.String
			analysis.GitRef = &value
		}
		if commitSHA.Valid {
			value := commitSHA.String
			analysis.CommitSHA = &value
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
		SELECT id, project_id, registry_id, image, tag, git_ref, commit_sha, status, total_size_bytes, layer_count, largest_layer_bytes, result_json, started_at, finished_at, analyzed_at, created_at, updated_at
		FROM image_analyses
		WHERE id = $1 AND project_id = $2
	`

	var analysis ImageAnalysis
	var registryID sql.NullString
	var totalSize sql.NullInt64
	var layerCount sql.NullInt64
	var largestLayer sql.NullInt64
	var gitRef sql.NullString
	var commitSHA sql.NullString
	var resultJSON sql.NullString
	var startedAt sql.NullTime
	var finishedAt sql.NullTime
	var analyzedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, query, analysisID, projectID).Scan(
		&analysis.ID,
		&analysis.ProjectID,
		&registryID,
		&analysis.Image,
		&analysis.Tag,
		&gitRef,
		&commitSHA,
		&analysis.Status,
		&totalSize,
		&layerCount,
		&largestLayer,
		&resultJSON,
		&startedAt,
		&finishedAt,
		&analyzedAt,
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
	if layerCount.Valid {
		value := int(layerCount.Int64)
		analysis.LayerCount = &value
	}
	if largestLayer.Valid {
		value := largestLayer.Int64
		analysis.LargestLayerBytes = &value
	}
	if resultJSON.Valid {
		analysis.ResultJSON = []byte(resultJSON.String)
	}
	if startedAt.Valid {
		value := startedAt.Time
		analysis.StartedAt = &value
	}
	if finishedAt.Valid {
		value := finishedAt.Time
		analysis.FinishedAt = &value
	}
	if analyzedAt.Valid {
		value := analyzedAt.Time
		analysis.AnalyzedAt = &value
	}
	if gitRef.Valid {
		value := gitRef.String
		analysis.GitRef = &value
	}
	if commitSHA.Valid {
		value := commitSHA.String
		analysis.CommitSHA = &value
	}

	return analysis, nil
}

func (r *Repository) DeleteAnalysis(ctx context.Context, projectID, analysisID uuid.UUID) error {
	const query = `
		DELETE FROM image_analyses
		WHERE id = $1 AND project_id = $2
	`

	result, err := r.db.ExecContext(ctx, query, analysisID, projectID)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrAnalysisNotFound
	}
	return nil
}

func (r *Repository) RerunAnalysis(ctx context.Context, projectID, analysisID uuid.UUID) error {
	const updateQuery = `
		UPDATE image_analyses
		SET status = $1,
			total_size_bytes = NULL,
			layer_count = NULL,
			largest_layer_bytes = NULL,
			result_json = NULL,
			started_at = NULL,
			finished_at = NULL,
			analyzed_at = NULL,
			updated_at = NOW()
		WHERE id = $2 AND project_id = $3
	`
	const jobQuery = `
		INSERT INTO analysis_jobs (analysis_id, status)
		VALUES ($1, $2)
	`

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	result, err := tx.ExecContext(ctx, updateQuery, StatusQueued, analysisID, projectID)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrAnalysisNotFound
	}

	if _, err = tx.ExecContext(ctx, jobQuery, analysisID, StatusQueued); err != nil {
		return err
	}

	return tx.Commit()
}

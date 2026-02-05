package analyses

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
)

func (r *Repository) GetAnalysisByID(ctx context.Context, analysisID uuid.UUID) (ImageAnalysis, error) {
	const query = `
		SELECT id, project_id, registry_id, image, tag, git_ref, commit_sha, status, total_size_bytes, layer_count, largest_layer_bytes,
		       result_json, started_at, finished_at, analyzed_at, created_at, updated_at
		FROM image_analyses
		WHERE id = $1
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

	err := r.db.QueryRowContext(ctx, query, analysisID).Scan(
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

func (r *Repository) GetLatestCompletedBaseline(ctx context.Context, projectID uuid.UUID, image, gitRef string, excludeID uuid.UUID) (ImageAnalysis, error) {
	const query = `
		SELECT id, project_id, registry_id, image, tag, git_ref, commit_sha, status, total_size_bytes, layer_count, largest_layer_bytes,
		       result_json, started_at, finished_at, analyzed_at, created_at, updated_at
		FROM image_analyses
		WHERE project_id = $1
		  AND image = $2
		  AND git_ref = $3
		  AND status = $4
		  AND id <> $5
		  AND analyzed_at IS NOT NULL
		ORDER BY analyzed_at DESC
		LIMIT 1
	`

	var analysis ImageAnalysis
	var registryID sql.NullString
	var totalSize sql.NullInt64
	var layerCount sql.NullInt64
	var largestLayer sql.NullInt64
	var gitRefOut sql.NullString
	var commitSHA sql.NullString
	var resultJSON sql.NullString
	var startedAt sql.NullTime
	var finishedAt sql.NullTime
	var analyzedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, projectID, image, gitRef, StatusCompleted, excludeID).Scan(
		&analysis.ID,
		&analysis.ProjectID,
		&registryID,
		&analysis.Image,
		&analysis.Tag,
		&gitRefOut,
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
		return ImageAnalysis{}, ErrBaselineNotFound
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
	if gitRefOut.Valid {
		value := gitRefOut.String
		analysis.GitRef = &value
	}
	if commitSHA.Valid {
		value := commitSHA.String
		analysis.CommitSHA = &value
	}

	return analysis, nil
}

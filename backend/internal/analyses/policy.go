package analyses

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

const (
	BaselineModeMainLatest = "main_latest"
	DefaultBaselineRef     = "main"
)

type Thresholds struct {
	WarnDeltaBytes  *int64
	FailDeltaBytes  *int64
	HardLimitBytes  *int64
	WarnDeltaLayers *int
	FailDeltaLayers *int
}

type ProjectPolicy struct {
	BaselineMode       string
	BaselineRefBranch  string
	BaselineAnalysisID *uuid.UUID
	Thresholds         Thresholds
}

func defaultProjectPolicy() ProjectPolicy {
	return ProjectPolicy{
		BaselineMode:      BaselineModeMainLatest,
		BaselineRefBranch: DefaultBaselineRef,
	}
}

func (r *Repository) GetProjectPolicy(ctx context.Context, projectID uuid.UUID) (ProjectPolicy, error) {
	const query = `
		SELECT baseline_mode, baseline_ref_branch, baseline_analysis_id,
		       warn_delta_bytes, fail_delta_bytes, hard_limit_bytes,
		       warn_delta_layers, fail_delta_layers
		FROM project_policies
		WHERE project_id = $1
	`

	policy := defaultProjectPolicy()
	var baselineAnalysisID sql.NullString
	var warnBytes, failBytes, hardLimit sql.NullInt64
	var warnLayers, failLayers sql.NullInt64
	var mode, refBranch sql.NullString

	err := r.db.QueryRowContext(ctx, query, projectID).Scan(
		&mode,
		&refBranch,
		&baselineAnalysisID,
		&warnBytes,
		&failBytes,
		&hardLimit,
		&warnLayers,
		&failLayers,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return policy, nil
		}
		return ProjectPolicy{}, err
	}

	if mode.Valid && mode.String != "" {
		policy.BaselineMode = mode.String
	}
	if refBranch.Valid && refBranch.String != "" {
		policy.BaselineRefBranch = refBranch.String
	}
	if baselineAnalysisID.Valid {
		parsed, err := uuid.Parse(baselineAnalysisID.String)
		if err != nil {
			return ProjectPolicy{}, err
		}
		policy.BaselineAnalysisID = &parsed
	}
	if warnBytes.Valid {
		value := warnBytes.Int64
		policy.Thresholds.WarnDeltaBytes = &value
	}
	if failBytes.Valid {
		value := failBytes.Int64
		policy.Thresholds.FailDeltaBytes = &value
	}
	if hardLimit.Valid {
		value := hardLimit.Int64
		policy.Thresholds.HardLimitBytes = &value
	}
	if warnLayers.Valid {
		value := int(warnLayers.Int64)
		policy.Thresholds.WarnDeltaLayers = &value
	}
	if failLayers.Valid {
		value := int(failLayers.Int64)
		policy.Thresholds.FailDeltaLayers = &value
	}

	return policy, nil
}

package analyses

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type HistoryFilter struct {
	Image  *string
	GitRef *string
	Status string
	From   *time.Time
	To     *time.Time
	Limit  int
}

type HistoryItem struct {
	ID                uuid.UUID
	Image             string
	GitRef            *string
	CommitSHA         *string
	Status            string
	AnalyzedAt        *time.Time
	TotalSizeBytes    *int64
	LayerCount        *int
	LargestLayerBytes *int64
}

type TrendMetric string

const (
	TrendMetricTotalSize    TrendMetric = "total_size_bytes"
	TrendMetricLayerCount   TrendMetric = "layer_count"
	TrendMetricLargestLayer TrendMetric = "largest_layer_bytes"
)

const (
	HistoryStatusAll      = "all"
	HistoryStatusQueued   = "queued"
	HistoryStatusRunning  = "running"
	HistoryStatusFailed   = "failed"
	HistoryStatusComplete = "completed"
)

type TrendPoint struct {
	Timestamp time.Time
	Value     int64
}

func (r *Repository) ListHistory(ctx context.Context, projectID uuid.UUID, filter HistoryFilter) ([]HistoryItem, error) {
	clauses := []string{"project_id = $1"}
	args := []any{projectID}
	idx := 2

	if filter.Image != nil {
		clauses = append(clauses, fmt.Sprintf("image = $%d", idx))
		args = append(args, *filter.Image)
		idx++
	}
	if filter.GitRef != nil {
		clauses = append(clauses, fmt.Sprintf("git_ref = $%d", idx))
		args = append(args, *filter.GitRef)
		idx++
	}
	if filter.Status != "" && filter.Status != HistoryStatusAll {
		clauses = append(clauses, fmt.Sprintf("status = $%d", idx))
		args = append(args, filter.Status)
		idx++
	}
	if filter.From != nil {
		clauses = append(clauses, fmt.Sprintf("analyzed_at >= $%d", idx))
		args = append(args, *filter.From)
		idx++
	}
	if filter.To != nil {
		clauses = append(clauses, fmt.Sprintf("analyzed_at <= $%d", idx))
		args = append(args, *filter.To)
		idx++
	}

	query := fmt.Sprintf(`
		SELECT id, image, git_ref, commit_sha, status, analyzed_at, total_size_bytes, layer_count, largest_layer_bytes
		FROM image_analyses
		WHERE %s
		ORDER BY analyzed_at DESC NULLS LAST
	`, strings.Join(clauses, " AND "))

	if filter.Limit > 0 {
		query = fmt.Sprintf("%s LIMIT $%d", query, idx)
		args = append(args, filter.Limit)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []HistoryItem{}
	for rows.Next() {
		var item HistoryItem
		var gitRef sql.NullString
		var commitSHA sql.NullString
		var analyzedAt sql.NullTime
		var totalSize sql.NullInt64
		var layerCount sql.NullInt64
		var largestLayer sql.NullInt64

		if err := rows.Scan(
			&item.ID,
			&item.Image,
			&gitRef,
			&commitSHA,
			&item.Status,
			&analyzedAt,
			&totalSize,
			&layerCount,
			&largestLayer,
		); err != nil {
			return nil, err
		}

		if gitRef.Valid {
			value := gitRef.String
			item.GitRef = &value
		}
		if commitSHA.Valid {
			value := commitSHA.String
			item.CommitSHA = &value
		}
		if analyzedAt.Valid {
			value := analyzedAt.Time
			item.AnalyzedAt = &value
		}
		if totalSize.Valid {
			value := totalSize.Int64
			item.TotalSizeBytes = &value
		}
		if layerCount.Valid {
			value := int(layerCount.Int64)
			item.LayerCount = &value
		}
		if largestLayer.Valid {
			value := largestLayer.Int64
			item.LargestLayerBytes = &value
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *Repository) ListTrends(ctx context.Context, projectID uuid.UUID, metric TrendMetric, filter HistoryFilter) ([]TrendPoint, error) {
	valueExpr := "total_size_bytes"
	valueFilter := "total_size_bytes IS NOT NULL"

	switch metric {
	case TrendMetricTotalSize:
		valueExpr = "total_size_bytes"
		valueFilter = "total_size_bytes IS NOT NULL"
	case TrendMetricLayerCount:
		valueExpr = "layer_count::bigint"
		valueFilter = "layer_count IS NOT NULL"
	case TrendMetricLargestLayer:
		valueExpr = "largest_layer_bytes"
		valueFilter = "largest_layer_bytes IS NOT NULL"
	default:
		return nil, fmt.Errorf("unsupported metric")
	}

	clauses := []string{"project_id = $1", "status = $2", "analyzed_at IS NOT NULL", valueFilter}
	args := []any{projectID, StatusCompleted}
	idx := 3

	if filter.Image != nil {
		clauses = append(clauses, fmt.Sprintf("image = $%d", idx))
		args = append(args, *filter.Image)
		idx++
	}
	if filter.GitRef != nil {
		clauses = append(clauses, fmt.Sprintf("git_ref = $%d", idx))
		args = append(args, *filter.GitRef)
		idx++
	}
	if filter.From != nil {
		clauses = append(clauses, fmt.Sprintf("analyzed_at >= $%d", idx))
		args = append(args, *filter.From)
		idx++
	}
	if filter.To != nil {
		clauses = append(clauses, fmt.Sprintf("analyzed_at <= $%d", idx))
		args = append(args, *filter.To)
		idx++
	}

	query := fmt.Sprintf(`
		SELECT analyzed_at, %s
		FROM image_analyses
		WHERE %s
		ORDER BY analyzed_at ASC
	`, valueExpr, strings.Join(clauses, " AND "))

	limit := filter.Limit
	if limit <= 0 || limit > 1000 {
		limit = 1000
	}
	query = fmt.Sprintf("%s LIMIT $%d", query, idx)
	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	points := []TrendPoint{}
	for rows.Next() {
		var ts time.Time
		var value sql.NullInt64
		if err := rows.Scan(&ts, &value); err != nil {
			return nil, err
		}
		if !value.Valid {
			continue
		}
		points = append(points, TrendPoint{Timestamp: ts, Value: value.Int64})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return points, nil
}

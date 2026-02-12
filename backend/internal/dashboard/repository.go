package dashboard

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetSummary(ctx context.Context, userID uuid.UUID) (Summary, error) {
	const query = `
		SELECT
			COUNT(DISTINCT pm.project_id) AS projects_total,
			COUNT(ia.id) AS analyses_total,
			COALESCE(SUM(CASE WHEN ia.status = 'completed' THEN 1 ELSE 0 END), 0) AS completed_total,
			COALESCE(SUM(CASE WHEN ia.status = 'running' THEN 1 ELSE 0 END), 0) AS running_total,
			COALESCE(SUM(CASE WHEN ia.status = 'failed' THEN 1 ELSE 0 END), 0) AS failed_total,
			COUNT(DISTINCT ia.image) FILTER (WHERE ia.image IS NOT NULL AND ia.image <> '') AS unique_images_total,
			COALESCE(SUM(CASE WHEN ia.created_at >= NOW() - INTERVAL '7 days' THEN 1 ELSE 0 END), 0) AS analyses_last_7_days,
			COALESCE(SUM(CASE WHEN ia.created_at >= NOW() - INTERVAL '30 days' THEN 1 ELSE 0 END), 0) AS analyses_last_30_days
		FROM project_members pm
		LEFT JOIN image_analyses ia ON ia.project_id = pm.project_id
		WHERE pm.user_id = $1
	`

	var summary Summary
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&summary.ProjectsTotal,
		&summary.AnalysesTotal,
		&summary.CompletedTotal,
		&summary.RunningTotal,
		&summary.FailedTotal,
		&summary.UniqueImagesTotal,
		&summary.AnalysesLast7Days,
		&summary.AnalysesLast30Days,
	)
	if err != nil {
		return Summary{}, err
	}

	return summary, nil
}

func (r *Repository) ListRecentEvents(ctx context.Context, userID uuid.UUID, limit int) ([]Event, error) {
	const query = `
		SELECT
			event_type,
			occurred_at,
			project_id,
			project_name,
			analysis_id,
			analysis_status,
			image,
			tag
		FROM (
			SELECT
				'project_created' AS event_type,
				p.created_at AS occurred_at,
				p.id AS project_id,
				p.name AS project_name,
				NULL::uuid AS analysis_id,
				NULL::text AS analysis_status,
				NULL::text AS image,
				NULL::text AS tag
			FROM projects p
			JOIN project_members pm ON pm.project_id = p.id
			WHERE pm.user_id = $1
			UNION ALL
			SELECT
				CASE
					WHEN ia.status = 'completed' THEN 'analysis_completed'
					WHEN ia.status = 'failed' THEN 'analysis_failed'
					WHEN ia.status = 'running' THEN 'analysis_running'
					ELSE 'analysis_queued'
				END AS event_type,
				ia.created_at AS occurred_at,
				p.id AS project_id,
				p.name AS project_name,
				ia.id AS analysis_id,
				ia.status AS analysis_status,
				ia.image AS image,
				ia.tag AS tag
			FROM image_analyses ia
			JOIN projects p ON p.id = ia.project_id
			JOIN project_members pm ON pm.project_id = ia.project_id
			WHERE pm.user_id = $1
		) events
		ORDER BY occurred_at DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]Event, 0, limit)
	for rows.Next() {
		var event Event
		var analysisID sql.NullString
		var analysisStatus sql.NullString
		var image sql.NullString
		var tag sql.NullString
		if err := rows.Scan(
			&event.Type,
			&event.OccurredAt,
			&event.ProjectID,
			&event.ProjectName,
			&analysisID,
			&analysisStatus,
			&image,
			&tag,
		); err != nil {
			return nil, err
		}
		if analysisID.Valid {
			parsedID, err := uuid.Parse(analysisID.String)
			if err != nil {
				return nil, err
			}
			event.AnalysisID = &parsedID
		}
		if analysisStatus.Valid {
			value := analysisStatus.String
			event.AnalysisStatus = &value
		}
		if image.Valid {
			value := image.String
			event.Image = &value
		}
		if tag.Valid {
			value := tag.String
			event.Tag = &value
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

func (r *Repository) ListDailyActivity(ctx context.Context, userID uuid.UUID, sinceDate time.Time) ([]DailyCount, error) {
	const query = `
		WITH contributions AS (
			SELECT DATE(p.created_at) AS day, COUNT(*)::int AS count
			FROM projects p
			JOIN project_members pm ON pm.project_id = p.id
			WHERE pm.user_id = $1
				AND p.created_at >= $2
			GROUP BY DATE(p.created_at)
			UNION ALL
			SELECT DATE(ia.created_at) AS day, COUNT(*)::int AS count
			FROM image_analyses ia
			JOIN project_members pm ON pm.project_id = ia.project_id
			WHERE pm.user_id = $1
				AND ia.created_at >= $2
			GROUP BY DATE(ia.created_at)
		)
		SELECT day, SUM(count)::int AS count
		FROM contributions
		GROUP BY day
		ORDER BY day ASC
	`

	rows, err := r.db.QueryContext(ctx, query, userID, sinceDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]DailyCount, 0, 32)
	for rows.Next() {
		var day DailyCount
		if err := rows.Scan(&day.Day, &day.Count); err != nil {
			return nil, err
		}
		result = append(result, day)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

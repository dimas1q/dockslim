package dashboard

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const (
	defaultRecentEventsLimit = 20
	defaultActivityDays      = 35
)

type RepositoryStore interface {
	GetSummary(ctx context.Context, userID uuid.UUID) (Summary, error)
	ListRecentEvents(ctx context.Context, userID uuid.UUID, limit int) ([]Event, error)
	ListDailyActivity(ctx context.Context, userID uuid.UUID, sinceDate time.Time) ([]DailyCount, error)
}

type Service struct {
	repo              RepositoryStore
	recentEventsLimit int
	activityDays      int
}

func NewService(repo RepositoryStore) *Service {
	return &Service{
		repo:              repo,
		recentEventsLimit: defaultRecentEventsLimit,
		activityDays:      defaultActivityDays,
	}
}

func (s *Service) GetDashboard(ctx context.Context, userID uuid.UUID) (AccountDashboard, error) {
	summary, err := s.repo.GetSummary(ctx, userID)
	if err != nil {
		return AccountDashboard{}, err
	}

	events, err := s.repo.ListRecentEvents(ctx, userID, s.recentEventsLimit)
	if err != nil {
		return AccountDashboard{}, err
	}

	now := time.Now().UTC()
	startDate := midnightUTC(now.AddDate(0, 0, -(s.activityDays - 1)))
	rawDaily, err := s.repo.ListDailyActivity(ctx, userID, startDate)
	if err != nil {
		return AccountDashboard{}, err
	}

	points := buildActivityPoints(startDate, s.activityDays, rawDaily)

	return AccountDashboard{
		Summary: summary,
		Activity: Activity{
			Last35Days:   points,
			RecentEvents: events,
		},
	}, nil
}

func buildActivityPoints(startDate time.Time, days int, rows []DailyCount) []ActivityPoint {
	byDay := make(map[string]int, len(rows))
	maxCount := 0
	for _, row := range rows {
		key := midnightUTC(row.Day).Format("2006-01-02")
		byDay[key] += row.Count
		if byDay[key] > maxCount {
			maxCount = byDay[key]
		}
	}

	points := make([]ActivityPoint, 0, days)
	for i := 0; i < days; i++ {
		day := midnightUTC(startDate.AddDate(0, 0, i))
		count := byDay[day.Format("2006-01-02")]
		points = append(points, ActivityPoint{
			Date:  day.Format("2006-01-02"),
			Count: count,
			Level: contributionLevel(count, maxCount),
		})
	}

	return points
}

func contributionLevel(count int, maxCount int) int {
	if count <= 0 {
		return 0
	}
	if maxCount <= 1 {
		return 4
	}

	ratio := float64(count) / float64(maxCount)
	switch {
	case ratio < 0.25:
		return 1
	case ratio < 0.5:
		return 2
	case ratio < 0.75:
		return 3
	default:
		return 4
	}
}

func midnightUTC(value time.Time) time.Time {
	utc := value.UTC()
	return time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
}

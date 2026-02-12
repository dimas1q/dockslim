package dashboard

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

type repoStub struct {
	summary Summary
	events  []Event
	daily   []DailyCount
}

func (s *repoStub) GetSummary(ctx context.Context, userID uuid.UUID) (Summary, error) {
	return s.summary, nil
}

func (s *repoStub) ListRecentEvents(ctx context.Context, userID uuid.UUID, limit int) ([]Event, error) {
	return s.events, nil
}

func (s *repoStub) ListDailyActivity(ctx context.Context, userID uuid.UUID, sinceDate time.Time) ([]DailyCount, error) {
	return s.daily, nil
}

func TestServiceBuildsDashboard(t *testing.T) {
	repo := &repoStub{
		summary: Summary{ProjectsTotal: 2, AnalysesTotal: 5},
		events: []Event{
			{Type: "analysis_completed", OccurredAt: time.Now().UTC(), ProjectID: uuid.New(), ProjectName: "core"},
		},
		daily: []DailyCount{
			{Day: time.Now().UTC().AddDate(0, 0, -1), Count: 1},
			{Day: time.Now().UTC().AddDate(0, 0, -2), Count: 4},
		},
	}

	svc := NewService(repo)
	out, err := svc.GetDashboard(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if out.Summary.ProjectsTotal != 2 {
		t.Fatalf("expected projects_total=2, got %d", out.Summary.ProjectsTotal)
	}
	if len(out.Activity.RecentEvents) != 1 {
		t.Fatalf("expected 1 recent event, got %d", len(out.Activity.RecentEvents))
	}
	if len(out.Activity.Last35Days) != 35 {
		t.Fatalf("expected 35 activity points, got %d", len(out.Activity.Last35Days))
	}
}

func TestContributionLevel(t *testing.T) {
	if got := contributionLevel(0, 5); got != 0 {
		t.Fatalf("expected level 0 for zero count, got %d", got)
	}
	if got := contributionLevel(1, 1); got != 4 {
		t.Fatalf("expected level 4 when max count is 1, got %d", got)
	}
	if got := contributionLevel(2, 8); got != 2 {
		t.Fatalf("expected level 2, got %d", got)
	}
}

package dashboard

import (
	"time"

	"github.com/google/uuid"
)

type Summary struct {
	ProjectsTotal      int `json:"projects_total"`
	AnalysesTotal      int `json:"analyses_total"`
	CompletedTotal     int `json:"completed_total"`
	RunningTotal       int `json:"running_total"`
	FailedTotal        int `json:"failed_total"`
	UniqueImagesTotal  int `json:"unique_images_total"`
	AnalysesLast7Days  int `json:"analyses_last_7_days"`
	AnalysesLast30Days int `json:"analyses_last_30_days"`
}

type Event struct {
	Type           string     `json:"type"`
	OccurredAt     time.Time  `json:"occurred_at"`
	ProjectID      uuid.UUID  `json:"project_id"`
	ProjectName    string     `json:"project_name"`
	AnalysisID     *uuid.UUID `json:"analysis_id,omitempty"`
	AnalysisStatus *string    `json:"analysis_status,omitempty"`
	Image          *string    `json:"image,omitempty"`
	Tag            *string    `json:"tag,omitempty"`
}

type DailyCount struct {
	Day   time.Time
	Count int
}

type ActivityPoint struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
	Level int    `json:"level"`
}

type Activity struct {
	Last35Days   []ActivityPoint `json:"last_35_days"`
	RecentEvents []Event         `json:"recent_events"`
}

type AccountDashboard struct {
	Summary  Summary  `json:"summary"`
	Activity Activity `json:"activity"`
}

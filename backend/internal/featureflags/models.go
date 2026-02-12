package featureflags

import (
	"time"

	"github.com/google/uuid"
)

const (
	PlanFree = "free"
	PlanPro  = "pro"
	PlanTeam = "team"

	SubscriptionStatusActive  = "active"
	SubscriptionStatusExpired = "expired"
)

const (
	FeatureBasicAnalysis    = "basic_analysis"
	FeatureAdvancedInsights = "advanced_insights"
	FeatureExportPDF        = "export_pdf"
	FeatureExportJSON       = "export_json"
	FeatureCIComments       = "ci_comments"
	FeatureBaselineSLA      = "baseline_sla"
	FeatureTeamManagement   = "team_management"
	FeatureSharedProjects   = "shared_projects"
	FeatureAdvancedTrends   = "advanced_trends"
	FeatureHistoryDaysLimit = "history_days_limit"
)

const (
	CICommentsModeLimited = "limited"
)

type Plan struct {
	ID        string
	Name      string
	Features  map[string]any
	CreatedAt time.Time
}

type UserSubscription struct {
	UserID     uuid.UUID
	PlanID     string
	Status     string
	ValidUntil *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type UpdateSubscriptionInput struct {
	UserID     uuid.UUID
	PlanID     string
	Status     string
	ValidUntil *time.Time
}

type UserFeatures struct {
	UserID     uuid.UUID
	PlanID     string
	PlanName   string
	Status     string
	ValidUntil *time.Time
	IsAdmin    bool
	Features   map[string]any
}

func (u UserFeatures) FeatureValue(featureName string) (any, bool) {
	if len(u.Features) == 0 {
		return nil, false
	}
	value, ok := u.Features[featureName]
	return value, ok
}

func (u UserFeatures) Limits() map[string]any {
	limits := make(map[string]any)
	for key, value := range u.Features {
		switch key {
		case FeatureHistoryDaysLimit, FeatureCIComments:
			limits[key] = value
		}
	}
	return limits
}

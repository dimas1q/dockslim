package budgets

import (
	"time"

	"github.com/google/uuid"
)

// Budget represents a budget configuration row.
type Budget struct {
	ID             uuid.UUID
	ProjectID      uuid.UUID
	Image          *string
	WarnDeltaBytes *int64
	FailDeltaBytes *int64
	HardLimitBytes *int64
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// ResolvedBudget contains thresholds picked for a specific image.
type ResolvedBudget struct {
	WarnDeltaBytes *int64
	FailDeltaBytes *int64
	HardLimitBytes *int64
}

// EvaluationResult captures budget verdict.
type EvaluationResult struct {
	Status         string   `json:"status"`
	Reasons        []string `json:"reasons"`
	DeltaBytes     int64    `json:"delta_bytes"`
	ToTotalBytes   int64    `json:"to_total_bytes"`
	WarnDeltaBytes *int64   `json:"warn_delta_bytes,omitempty"`
	FailDeltaBytes *int64   `json:"fail_delta_bytes,omitempty"`
	HardLimitBytes *int64   `json:"hard_limit_bytes,omitempty"`
}

package analyses

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const (
	StatusQueued    = "queued"
	StatusRunning   = "running"
	StatusCompleted = "completed"
	StatusFailed    = "failed"
)

type ImageAnalysis struct {
	ID                uuid.UUID
	ProjectID         uuid.UUID
	RegistryID        *uuid.UUID
	Image             string
	Tag               string
	GitRef            *string
	CommitSHA         *string
	Status            string
	TotalSizeBytes    *int64
	LayerCount        *int
	LargestLayerBytes *int64
	ResultJSON        json.RawMessage
	StartedAt         *time.Time
	FinishedAt        *time.Time
	AnalyzedAt        *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

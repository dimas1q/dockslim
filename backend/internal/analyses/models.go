package analyses

import (
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
	ID             uuid.UUID
	ProjectID      uuid.UUID
	RegistryID     *uuid.UUID
	Image          string
	Tag            string
	Status         string
	TotalSizeBytes *int64
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

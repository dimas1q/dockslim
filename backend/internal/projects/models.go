package projects

import (
	"time"

	"github.com/google/uuid"
)

const RoleOwner = "owner"

// Project represents a project owned by one or more users.
type Project struct {
	ID          uuid.UUID
	Name        string
	Description *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

package citokens

import (
	"time"

	"github.com/google/uuid"
)

const TokenPrefix = "ds_ci_"

// Token represents a project-scoped CI token.
type Token struct {
	ID         uuid.UUID
	ProjectID  uuid.UUID
	Name       string
	TokenHash  string
	LastUsedAt *time.Time
	CreatedAt  time.Time
	RevokedAt  *time.Time
	ExpiresAt  *time.Time
}

package apitokens

import (
	"time"

	"github.com/google/uuid"
)

const TokenPrefix = "ds_api_"

// Token represents a user-scoped API token.
type Token struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	Name       string
	TokenHash  string
	LastUsedAt *time.Time
	CreatedAt  time.Time
	RevokedAt  *time.Time
	ExpiresAt  *time.Time
}

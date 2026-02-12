package auth

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	Login        string
	Email        string
	PasswordHash string
	IsAdmin      bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type AuthKey struct {
	ID         uuid.UUID
	KeyID      string
	SigningKey string
	Algorithm  string
	IsActive   bool
	CreatedAt  time.Time
	RotatedAt  *time.Time
}

type RefreshToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

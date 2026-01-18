package registries

import (
	"time"

	"github.com/google/uuid"
)

const (
	TypeGeneric = "generic"
)

type Registry struct {
	ID          uuid.UUID
	ProjectID   uuid.UUID
	Name        string
	Type        string
	RegistryURL string
	Username    *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type EncryptionKey struct {
	ID          uuid.UUID
	KeyID       string
	KeyMaterial []byte
	IsActive    bool
	CreatedAt   time.Time
}

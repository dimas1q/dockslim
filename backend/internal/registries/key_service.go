package registries

import (
	"context"
	"crypto/rand"
	"errors"

	"github.com/google/uuid"
)

type KeyRepository interface {
	GetActiveKey(ctx context.Context) (EncryptionKey, error)
	CreateKey(ctx context.Context, key EncryptionKey) (EncryptionKey, error)
}

func EnsureActiveKey(ctx context.Context, repo KeyRepository) (EncryptionKey, error) {
	key, err := repo.GetActiveKey(ctx)
	if err == nil {
		return key, nil
	}
	if !errors.Is(err, ErrKeyNotFound) {
		return EncryptionKey{}, err
	}

	material := make([]byte, 32)
	if _, err := rand.Read(material); err != nil {
		return EncryptionKey{}, err
	}

	newKey := EncryptionKey{
		KeyID:       uuid.NewString(),
		KeyMaterial: material,
		IsActive:    true,
	}

	return repo.CreateKey(ctx, newKey)
}

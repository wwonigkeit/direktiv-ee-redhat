package datastore

import (
	"context"
	"crypto/sha256"
	"time"

	"github.com/google/uuid"
)

type APIToken struct {
	Name        string
	Namespace   string
	Description string
	Hash        uuid.UUID
	Permissions Permissions
	ExpiredAt   time.Time
	IsExpired   bool

	CreatedAt time.Time
	UpdatedAt time.Time
}

type APITokensStore interface {
	Create(ctx context.Context, apiToken *APIToken, lifeSeconds int) (*APIToken, error)
	Delete(ctx context.Context, namespace, name string) error
	Get(ctx context.Context, namespace, name string) (*APIToken, error)
	GetByHash(ctx context.Context, hash uuid.UUID) (*APIToken, error)
	List(ctx context.Context, namespace string) ([]*APIToken, error)
}

func HashTokenID(input uuid.UUID) uuid.UUID {
	sum := sha256.Sum256(input[:])
	output := make([]byte, 0, 16)

	output = append(output, input[0:10]...)
	output = append(output, sum[0:6]...)

	return uuid.UUID(output)
}

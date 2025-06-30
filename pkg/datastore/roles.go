package datastore

import (
	"context"
	"time"
)

type Role struct {
	Name        string
	Namespace   string
	Description string
	OidcGroups  OidcGroups
	Permissions Permissions

	CreatedAt time.Time
	UpdatedAt time.Time
}

type RolesStore interface {
	Create(ctx context.Context, role *Role) (*Role, error)
	Delete(ctx context.Context, namespace, name string) error
	Get(ctx context.Context, namespace, name string) (*Role, error)
	Update(ctx context.Context, namespace, name string, role *Role) (*Role, error)
	List(ctx context.Context, namespace string) ([]*Role, error)
	ListAll(ctx context.Context) ([]*Role, error)
}

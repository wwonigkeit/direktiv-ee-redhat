package datasql

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/direktiv/direktiv/direktiv-ee/pkg/datastore"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type apiTokensStore struct {
	db *gorm.DB
}

//nolint:goconst
func (s *apiTokensStore) Create(ctx context.Context, apiToken *datastore.APIToken, lifeSeconds int) (*datastore.APIToken, error) {
	vErrs := datastore.InvalidArgumentError{}
	if apiToken == nil {
		vErrs["apiToken"] = "is nil"

		return nil, vErrs
	}
	if apiToken.Name == "" {
		vErrs["name"] = "is required"
	}
	if apiToken.Namespace == "" {
		vErrs["namespace"] = "is required"
	}
	if apiToken.Hash == uuid.Nil {
		vErrs["hash"] = "is required"
	}
	err := apiToken.Permissions.Validate()
	if err != nil {
		vErrs["permissions"] = err.Error()
	}
	if len(vErrs) > 0 {
		return nil, vErrs
	}
	for i := range apiToken.Permissions {
		apiToken.Permissions[i].Namespace = apiToken.Namespace
	}
	query := fmt.Sprintf(`
							INSERT INTO ee_api_tokens(name, namespace, description, hash, permissions, expired_at) VALUES(?, ?, ?, ?, ?, NOW() + INTERVAL '%d SECOND');
							`, lifeSeconds)

	res := s.db.WithContext(ctx).Exec(query, apiToken.Name, apiToken.Namespace, apiToken.Description, apiToken.Hash, apiToken.Permissions)

	if res.Error != nil && strings.Contains(res.Error.Error(), "duplicate key") {
		return nil, datastore.ErrDuplication
	}
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected != 1 {
		return nil, fmt.Errorf("unexpected ee_api_tokens insert count, got: %d, want: %d", res.RowsAffected, 1)
	}

	return s.Get(ctx, apiToken.Namespace, apiToken.Name)
}

func (s *apiTokensStore) Delete(ctx context.Context, namespace, name string) error {
	res := s.db.WithContext(ctx).Exec(`DELETE FROM ee_api_tokens WHERE  name=? AND namespace=?`, name, namespace)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected != 1 {
		return datastore.ErrNotFound
	}

	return nil
}

func (s *apiTokensStore) Get(ctx context.Context, namespace, name string) (*datastore.APIToken, error) {
	scan := &datastore.APIToken{}
	res := s.db.WithContext(ctx).Raw(`
							SELECT name, namespace, description, hash, permissions, expired_at, created_at, updated_at,
							(expired_at <= NOW()) AS is_expired
							FROM ee_api_tokens 
							WHERE name=? AND namespace=?`,
		name, namespace).
		First(scan)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, datastore.ErrNotFound
	}
	if res.Error != nil {
		return nil, res.Error
	}

	return scan, nil
}

func (s *apiTokensStore) GetByHash(ctx context.Context, hash uuid.UUID) (*datastore.APIToken, error) {
	scan := &datastore.APIToken{}
	res := s.db.WithContext(ctx).Raw(`
							SELECT name, namespace, description, hash, permissions, expired_at, created_at, updated_at,
							(expired_at <= NOW()) AS is_expired
							FROM ee_api_tokens 
							WHERE hash=?`,
		hash).
		First(scan)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, datastore.ErrNotFound
	}
	if res.Error != nil {
		return nil, res.Error
	}

	return scan, nil
}

func (s *apiTokensStore) List(ctx context.Context, namespace string) ([]*datastore.APIToken, error) {
	var list []*datastore.APIToken
	res := s.db.WithContext(ctx).Raw(`
							SELECT name, namespace, description, hash, permissions, expired_at, created_at, updated_at,
							(expired_at <= NOW()) AS is_expired
							FROM ee_api_tokens
							WHERE namespace=? 
							ORDER BY created_at ASC`, namespace).
		Find(&list)
	if res.Error != nil {
		return nil, res.Error
	}

	return list, nil
}

var _ datastore.APITokensStore = &apiTokensStore{}

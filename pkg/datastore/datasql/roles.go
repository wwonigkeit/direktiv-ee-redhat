package datasql

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/direktiv/direktiv/direktiv-ee/pkg/datastore"
	"gorm.io/gorm"
)

type rolesStore struct {
	db *gorm.DB
}

func (s *rolesStore) Update(ctx context.Context, namespace, name string, role *datastore.Role) (*datastore.Role, error) {
	vErrs := datastore.InvalidArgumentError{}
	if namespace == "" {
		//nolint:goconst
		vErrs["namespace"] = "is required"
	}
	if name == "" {
		vErrs["name"] = "is required"
	}
	if role == nil {
		//nolint:goconst
		vErrs["role"] = "is nil"

		return nil, vErrs
	}
	if role.Name == "" {
		vErrs["name"] = "is required"
	}
	err := role.Permissions.Validate()
	if err != nil {
		vErrs["permissions"] = err.Error()
	}
	err = role.OidcGroups.Validate()
	if err != nil {
		vErrs["oidcGroups"] = err.Error()
	}
	if len(vErrs) > 0 {
		return nil, vErrs
	}

	for i := range role.Permissions {
		role.Permissions[i].Namespace = role.Namespace
	}
	res := s.db.WithContext(ctx).Exec(`UPDATE ee_roles SET name=?, description=?, oidc_groups=?, permissions=?, updated_at=CURRENT_TIMESTAMP WHERE namespace=? and name=?`,
		role.Name, role.Description, role.OidcGroups, role.Permissions, namespace, name)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected != 1 {
		return nil, datastore.ErrNotFound
	}

	return s.Get(ctx, namespace, role.Name)
}

func (s *rolesStore) Create(ctx context.Context, role *datastore.Role) (*datastore.Role, error) {
	vErrs := datastore.InvalidArgumentError{}
	if role == nil {
		vErrs["role"] = "is nil"

		return nil, vErrs
	}
	if role.Namespace == "" {
		vErrs["namespace"] = "is required"
	}
	if role.Name == "" {
		vErrs["name"] = "is required"
	}
	err := role.Permissions.Validate()
	if err != nil {
		vErrs["permissions"] = err.Error()
	}
	err = role.OidcGroups.Validate()
	if err != nil {
		vErrs["oidcGroups"] = err.Error()
	}
	if len(vErrs) > 0 {
		return nil, vErrs
	}
	for i := range role.Permissions {
		role.Permissions[i].Namespace = role.Namespace
	}

	res := s.db.WithContext(ctx).Exec(`
							INSERT INTO ee_roles(name, namespace, description, oidc_groups, permissions) VALUES(?, ?, ?, ?, ?);
							`, role.Name, role.Namespace, role.Description, role.OidcGroups, role.Permissions)

	if res.Error != nil && strings.Contains(res.Error.Error(), "duplicate key") {
		return nil, datastore.ErrDuplication
	}
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected != 1 {
		return nil, fmt.Errorf("unexpected ee_roles insert count, got: %d, want: %d", res.RowsAffected, 1)
	}

	return s.Get(ctx, role.Namespace, role.Name)
}

func (s *rolesStore) Delete(ctx context.Context, namespace, name string) error {
	res := s.db.WithContext(ctx).Exec(`DELETE FROM ee_roles WHERE  name=? AND namespace=?`, name, namespace)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected != 1 {
		return datastore.ErrNotFound
	}

	return nil
}

func (s *rolesStore) Get(ctx context.Context, namespace, name string) (*datastore.Role, error) {
	scan := &datastore.Role{}
	res := s.db.WithContext(ctx).Raw(`
							SELECT name, namespace, description, oidc_groups, permissions, created_at, updated_at 
							FROM ee_roles 
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

func (s *rolesStore) List(ctx context.Context, namespace string) ([]*datastore.Role, error) {
	var list []*datastore.Role

	res := s.db.WithContext(ctx).Raw(`
							SELECT name, namespace, description, oidc_groups, permissions, created_at, updated_at 
							FROM ee_roles
							WHERE namespace=? 
							ORDER BY created_at ASC`, namespace).
		Find(&list)
	if res.Error != nil {
		return nil, res.Error
	}

	return list, nil
}

func (s *rolesStore) ListAll(ctx context.Context) ([]*datastore.Role, error) {
	var list []*datastore.Role

	res := s.db.WithContext(ctx).Raw(`
							SELECT name, namespace, description, oidc_groups, permissions, created_at, updated_at 
							FROM ee_roles
							ORDER BY created_at ASC`).
		Find(&list)
	if res.Error != nil {
		return nil, res.Error
	}

	return list, nil
}

var _ datastore.RolesStore = &rolesStore{}

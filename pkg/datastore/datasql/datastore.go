package datasql

import (
	_ "embed"
	"fmt"

	"github.com/direktiv/direktiv/direktiv-ee/pkg/datastore"
	"gorm.io/gorm"
)

//go:embed db_schema.sql
var Schema string

type store struct{}

var _ datastore.Store = &store{}

func (s *store) With(db any) datastore.StoreInner {
	gormDB, ok := db.(*gorm.DB)
	if !ok {
		panic(fmt.Sprintf("type error: invalid db type for sql store, want: *gorm.DB, got: %T", db))
	}

	return &storeInner{
		db: gormDB,
	}
}

func New() datastore.Store {
	return &store{}
}

type storeInner struct {
	db *gorm.DB
}

func (s *storeInner) APITokens() datastore.APITokensStore {
	return &apiTokensStore{db: s.db}
}

func (s *storeInner) Roles() datastore.RolesStore {
	return &rolesStore{db: s.db}
}

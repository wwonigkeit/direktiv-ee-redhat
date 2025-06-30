package datasql_test

import (
	"context"
	"errors"
	"github.com/direktiv/direktiv/direktiv-ee/pkg/datastore"
	"github.com/direktiv/direktiv/direktiv-ee/pkg/datastore/datasql"
	"github.com/direktiv/direktiv/pkg/database"
	"github.com/google/uuid"
	"testing"
)

const (
	textSomething     = "something"
	textSomethingElse = "something_else"
)

func Test_APITokens(t *testing.T) {
	ctx := context.Background()

	db, ns, err := database.NewTestDBWithNamespace(t, uuid.NewString())
	if err != nil {
		t.Fatalf("unepxected NewTestDBWithNamespace() error = %v", err)
	}
	if res := db.Conn().Exec(datasql.Schema); res.Error != nil {
		t.Fatalf("unepxected exec db_schema error = %v", res.Error)
	}

	res, err := datasql.New().With(db.Conn()).APITokens().Get(ctx, textSomething, textSomethingElse)
	if res != nil {
		t.Errorf("APITokens().Get() returned %v, want nil", res)
	}
	if !errors.Is(err, datastore.ErrNotFound) {
		t.Errorf("APITokens().Get() error = %v, wantErr %v", err, datastore.ErrNotFound)
	}

	err = datasql.New().With(db.Conn()).APITokens().Delete(ctx, textSomething, textSomethingElse)
	if !errors.Is(err, datastore.ErrNotFound) {
		t.Errorf("APITokens().Delete() error = %v, wantErr %v", err, datastore.ErrNotFound)
	}

	uuid1 := uuid.New()
	p1, err := datasql.New().With(db.Conn()).APITokens().Create(ctx, &datastore.APIToken{
		Name:        textSomething,
		Description: textSomethingElse,
		Namespace:   ns.Name,
		Hash:        uuid1,
		Permissions: datastore.Permissions{
			{"", "secrets", "GET"},
			{"", "variables", "GET"},
		},
	}, 0)
	if err != nil {
		t.Fatalf("APITokens().Create() error = %v", err)
	}
	if p1 == nil {
		t.Fatalf("APITokens().Create() returned nil")
	}
	if p1.Name != textSomething {
		t.Errorf("APITokens().Create() returned %v, want %v", p1.Name, textSomething)
	}
	if p1.Description != textSomethingElse {
		t.Errorf("APITokens().Create() returned %v, want %v", p1.Description, textSomethingElse)
	}
	if p1.Namespace != ns.Name {
		t.Errorf("APITokens().Create() returned %v, want %v", p1.Namespace, ns.Name)
	}
	if p1.Hash != uuid1 {
		t.Errorf("APITokens().Create() returned %v, want %v", p1.Hash, uuid1)
	}
	if p1.Permissions[0].Topic != "secrets" {
		t.Errorf("APITokens().Create() returned %v, want %v", p1.Permissions, "secrets")
	}

	l, err := datasql.New().With(db.Conn()).APITokens().List(ctx, ns.Name)
	if err != nil {
		t.Fatalf("APITokens().List() error = %v", err)
	}
	if len(l) != 1 {
		t.Errorf("APITokens().List() returned %v, want %v", len(l), 1)
	}
	if l[0].Name != textSomething {
		t.Errorf("APITokens().List() returned %v, want %v", l[0].Name, textSomething)
	}
}

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

func Test_Roles(t *testing.T) {
	ctx := context.Background()

	db, ns, err := database.NewTestDBWithNamespace(t, uuid.NewString())
	if err != nil {
		t.Fatalf("unepxected NewTestDBWithNamespace() error = %v", err)
	}
	if res := db.Conn().Exec(datasql.Schema); res.Error != nil {
		t.Fatalf("unepxected exec db_schema error = %v", res.Error)
	}

	res, err := datasql.New().With(db.Conn()).Roles().Get(ctx, textSomething, textSomethingElse)
	if res != nil {
		t.Errorf("Roles().Get() returned %v, want nil", res)
	}
	if !errors.Is(err, datastore.ErrNotFound) {
		t.Errorf("Roles().Get() error = %v, wantErr %v", err, datastore.ErrNotFound)
	}

	err = datasql.New().With(db.Conn()).Roles().Delete(ctx, textSomething, textSomethingElse)
	if !errors.Is(err, datastore.ErrNotFound) {
		t.Errorf("Roles().Delete() error = %v, wantErr %v", err, datastore.ErrNotFound)
	}

	p1, err := datasql.New().With(db.Conn()).Roles().Create(ctx, &datastore.Role{
		Name:        textSomething,
		Description: textSomethingElse,
		Namespace:   ns.Name,
		OidcGroups:  []string{"g1", "g2"},
		Permissions: datastore.Permissions{
			{"", "secrets", "GET"},
			{"", "variables", "POST"},
		},
	})
	if err != nil {
		t.Fatalf("Roles().Create() error = %v", err)
	}
	if p1 == nil {
		t.Fatalf("Roles().Create() returned nil")
	}
	if p1.Name != textSomething {
		t.Errorf("Roles().Create() returned %v, want %v", p1.Name, textSomething)
	}
	if p1.Description != textSomethingElse {
		t.Errorf("Roles().Create() returned %v, want %v", p1.Description, textSomethingElse)
	}
	if p1.Namespace != ns.Name {
		t.Errorf("Roles().Create() returned %v, want %v", p1.Namespace, ns.Name)
	}
	if p1.OidcGroups.String() != "[\"g1\",\"g2\"]" {
		t.Errorf("Roles().Create() returned %v, want %v", p1.OidcGroups.String(), "[\"g1\",\"g2\"]")
	}
	if p1.Permissions[0].Topic != "secrets" {
		t.Errorf("Roles().Create() returned %v, want %v", p1.Permissions, "secrets")
	}

	l, err := datasql.New().With(db.Conn()).Roles().List(ctx, ns.Name)
	if err != nil {
		t.Fatalf("Roles().List() error = %v", err)
	}
	if len(l) != 1 {
		t.Errorf("Roles().List() returned %v, want %v", len(l), 1)
	}
	if l[0].Name != textSomething {
		t.Errorf("Roles().List() returned %v, want %v", l[0].Name, textSomething)
	}
	p1, err = datasql.New().With(db.Conn()).Roles().Update(ctx, ns.Name, textSomething, &datastore.Role{
		Name:        textSomething,
		Description: textSomethingElse,
		Namespace:   ns.Name,
		OidcGroups:  []string{"g1"},
		Permissions: datastore.Permissions{
			{"", "secrets", "GET"},
			{"", "variables", "GET"},
		},
	})
	if err != nil {
		t.Fatalf("Roles().Update() error = %v", err)
	}
	if p1 == nil {
		t.Fatalf("Roles().Update() returned nil")
	}
	if p1.Name != textSomething {
		t.Errorf("Roles().Update() returned %v, want %v", p1.Name, textSomething)
	}
	if p1.OidcGroups.String() != "[\"g1\"]" {
		t.Errorf("Roles().Update() returned %v, want %v", p1.OidcGroups, "[\"g1\"]")
	}
	if p1.Permissions[1].Method != "GET" {
		t.Errorf("Roles().Update() returned %v, want %v", p1.Permissions, "GET")
	}
}

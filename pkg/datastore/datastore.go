package datastore

import (
	"errors"
	"fmt"
)

type Store interface {
	With(db any) StoreInner
}

type StoreInner interface {
	APITokens() APITokensStore
	Roles() RolesStore
}

var (
	// ErrNotFound is a common error type that should be returned by any store implementation
	// for the error cases when getting a single entry failed due to none existence.
	ErrNotFound = errors.New("not found")

	// ErrDuplication is a common error type that should be returned by any store implementation
	// when tying to violate unique constraints.
	ErrDuplication = errors.New("duplicate key")
)

type InvalidArgumentError map[string]string

func (e InvalidArgumentError) Error() string {
	str := ""
	for k, v := range e {
		str += fmt.Sprintf("field:'%s', err:%s,", k, v)
	}

	return "validation errors:" + str[:len(str)-1]
}

package datastore

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"slices"
)

var allowedMethods = []string{
	"POST",
	"GET",
	"DELETE",
	"PATCH",
	"PUT",
	"read",
	"manage",
}

var allowedTopics = []string{
	"namespaces",
	"instances",
	"syncs",
	"secrets",
	"variables",
	"files",
	"services",
	"registries",
	"logs",
	"notifications",
	"metrics",
	"events",
	"roles",
	"api_tokens",
}

type Permission struct {
	Namespace string
	Topic     string
	Method    string
}

//nolint:recvcheck
type Permissions []*Permission

func (perms Permissions) Validate() error {
	if len(perms) == 0 {
		return nil
	}

	for _, perm := range perms {
		if !slices.Contains(allowedMethods, perm.Method) {
			return fmt.Errorf("invalid permission method: '%s'", perm.Method)
		}
		if !slices.Contains(allowedTopics, perm.Topic) {
			return fmt.Errorf("invalid permission topic: '%s'", perm.Topic)
		}
	}

	return nil
}

func (perms Permissions) Value() (driver.Value, error) {
	return json.Marshal(perms)
}

func (perms Permissions) String() string {
	b, err := json.Marshal(perms)
	if err != nil {
		return ""
	}

	return string(b)
}

func (perms *Permissions) Scan(value interface{}) error {
	b, ok := value.(string)
	if !ok {
		return fmt.Errorf("type assertion to string failed: got %T", value)
	}
	if b == "" {
		return nil
	}

	return json.Unmarshal([]byte(b), perms)
}

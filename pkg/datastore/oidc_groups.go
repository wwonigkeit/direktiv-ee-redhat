package datastore

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

//nolint:recvcheck
type OidcGroups []string

func (g OidcGroups) Validate() error {
	if len(g) == 0 {
		return nil
	}

	for _, i := range g {
		if len(i) == 0 {
			return fmt.Errorf("empty OidcGroup string: '%s'", i)
		}
	}

	return nil
}

func (g OidcGroups) Value() (driver.Value, error) {
	return json.Marshal(g)
}

func (g OidcGroups) String() string {
	b, err := json.Marshal(g)
	if err != nil {
		return ""
	}

	return string(b)
}

func (g *OidcGroups) Scan(value interface{}) error {
	b, ok := value.(string)
	if !ok {
		return fmt.Errorf("type assertion to string failed: got %T", value)
	}
	if b == "" {
		return nil
	}

	return json.Unmarshal([]byte(b), g)
}

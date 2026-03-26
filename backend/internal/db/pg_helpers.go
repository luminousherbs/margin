package db

import (
	"database/sql/driver"
	"fmt"
	"strings"
)

type pqStringArray []string

func (a pqStringArray) Value() (driver.Value, error) {
	if a == nil {
		return "{}", nil
	}
	parts := make([]string, len(a))
	for i, s := range a {
		escaped := strings.ReplaceAll(s, "\\", "\\\\")
		escaped = strings.ReplaceAll(escaped, "\"", "\\\"")
		parts[i] = fmt.Sprintf(`"%s"`, escaped)
	}
	return "{" + strings.Join(parts, ",") + "}", nil
}

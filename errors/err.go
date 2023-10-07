package errors

import (
	"fmt"
)

func MissingSQL(id string) error {
	return fmt.Errorf("missing sql: %s", id)
}

func MissingParameter(propName string) error {
	return fmt.Errorf("missing parameter: %s", propName)
}

func TmplExecute(err error) error {
	return fmt.Errorf("failed execute template: %w", err)
}

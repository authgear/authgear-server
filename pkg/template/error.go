package template

import (
	"errors"
	"fmt"
)

type errNotFound struct {
	name string
}

func (e *errNotFound) Error() string { return fmt.Sprintf("template: '%s' not found", e.name) }

func IsNotFound(e error) bool {
	var err *errNotFound
	return errors.As(e, &err)
}

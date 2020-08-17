package template

import (
	"errors"
	"fmt"
)

var ErrInvalidDataURI = errors.New("template: invalid data URI")
var ErrInvalidUTF8 = errors.New("template: expected content to be UTF-8 encoded")
var ErrLimitReached = errors.New("template: rendered template is too large")
var ErrNoLanguageMatch = errors.New("template: no language match")

type errNotFound struct {
	name string
}

func (e *errNotFound) Error() string { return fmt.Sprintf("template: '%s' not found", e.name) }

func IsNotFound(e error) bool {
	var err *errNotFound
	return errors.As(e, &err)
}

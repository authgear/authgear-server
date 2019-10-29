package store

import "errors"

type errNotFound struct {
	name string
}

func (e *errNotFound) Error() string { return e.name + " not found" }

func NewNotFoundError(name string) error { return &errNotFound{name} }
func IsNotFound(e error) bool {
	var err *errNotFound
	return errors.As(e, &err)
}

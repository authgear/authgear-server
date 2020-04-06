package webapp

import (
	"errors"
)

var ErrStateNotFound = errors.New("state not found")

type StateStore interface {
	Get(id string) (*State, error)
	Set(state *State) error
}

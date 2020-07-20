package newinteraction

import "github.com/authgear/authgear-server/pkg/core/errors"

var ErrInvalidState = errors.New("invalid state")

type Store struct {
}

func (s *Store) Create(graph *Graph) error {
	return nil
}

func (s *Store) Delete(instanceID string) error {
	return nil
}

func (s *Store) Get(instanceID string) (*Graph, error) {
	return nil, nil
}

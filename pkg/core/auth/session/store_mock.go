package session

import (
	"fmt"

	"github.com/skygeario/skygear-server/pkg/core/auth"
)

type MockStore struct {
	Sessions map[string]auth.Session
}

var _ Store = &MockStore{}

func NewMockStore() *MockStore {
	return &MockStore{
		Sessions: map[string]auth.Session{},
	}
}

func (s *MockStore) Create(sess *auth.Session) error {
	if _, exists := s.Sessions[sess.ID]; exists {
		return fmt.Errorf("cannot create session")
	}
	s.Sessions[sess.ID] = *sess
	return nil
}

func (s *MockStore) Update(sess *auth.Session) error {
	if _, exists := s.Sessions[sess.ID]; !exists {
		return ErrSessionNotFound
	}
	s.Sessions[sess.ID] = *sess
	return nil
}

func (s *MockStore) Get(id string) (*auth.Session, error) {
	sess, exists := s.Sessions[id]
	if !exists {
		return nil, ErrSessionNotFound
	}
	return &sess, nil
}

func (s *MockStore) Delete(id string) error {
	delete(s.Sessions, id)
	return nil
}

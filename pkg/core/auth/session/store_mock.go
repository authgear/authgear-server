package session

import (
	"fmt"
	"sort"
	"time"

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

func (s *MockStore) Create(sess *auth.Session, expireAt time.Time) error {
	if _, exists := s.Sessions[sess.ID]; exists {
		return fmt.Errorf("cannot create session")
	}
	s.Sessions[sess.ID] = *sess
	return nil
}

func (s *MockStore) Update(sess *auth.Session, expireAt time.Time) error {
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

func (s *MockStore) List(userID string) (sessions []*auth.Session, err error) {
	for _, session := range s.Sessions {
		if session.UserID == userID {
			s := session
			sessions = append(sessions, &s)
		}
	}
	sort.Sort(sessionSlice(sessions))
	return
}

type sessionSlice []*auth.Session

func (s sessionSlice) Len() int           { return len(s) }
func (s sessionSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s sessionSlice) Less(i, j int) bool { return s[i].CreatedAt.Before(s[j].CreatedAt) }

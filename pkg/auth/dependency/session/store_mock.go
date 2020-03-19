package session

import (
	"fmt"
	"sort"
	"time"
)

type MockStore struct {
	Sessions map[string]IDPSession
}

var _ Store = &MockStore{}

func NewMockStore() *MockStore {
	return &MockStore{
		Sessions: map[string]IDPSession{},
	}
}

func (s *MockStore) Create(sess *IDPSession, expireAt time.Time) error {
	if _, exists := s.Sessions[sess.ID]; exists {
		return fmt.Errorf("cannot create session")
	}
	s.Sessions[sess.ID] = *sess
	return nil
}

func (s *MockStore) Update(sess *IDPSession, expireAt time.Time) error {
	if _, exists := s.Sessions[sess.ID]; !exists {
		return ErrSessionNotFound
	}
	s.Sessions[sess.ID] = *sess
	return nil
}

func (s *MockStore) Get(id string) (*IDPSession, error) {
	sess, exists := s.Sessions[id]
	if !exists {
		return nil, ErrSessionNotFound
	}
	return &sess, nil
}

func (s *MockStore) Delete(session *IDPSession) error {
	delete(s.Sessions, session.ID)
	return nil
}

func (s *MockStore) DeleteBatch(sessions []*IDPSession) error {
	for _, session := range sessions {
		delete(s.Sessions, session.ID)
	}
	return nil
}

func (s *MockStore) DeleteAll(userID string, sessionID string) error {
	for _, session := range s.Sessions {
		if session.Attrs.UserID == userID && session.ID != sessionID {
			delete(s.Sessions, session.ID)
		}
	}
	return nil
}

func (s *MockStore) List(userID string) (sessions []*IDPSession, err error) {
	for _, session := range s.Sessions {
		if session.Attrs.UserID == userID {
			s := session
			sessions = append(sessions, &s)
		}
	}
	sort.Sort(sessionSlice(sessions))
	return
}

type sessionSlice []*IDPSession

func (s sessionSlice) Len() int           { return len(s) }
func (s sessionSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s sessionSlice) Less(i, j int) bool { return s[i].CreatedAt.Before(s[j].CreatedAt) }

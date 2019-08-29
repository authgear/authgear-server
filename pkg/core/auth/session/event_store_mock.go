package session

import (
	"github.com/skygeario/skygear-server/pkg/core/auth"
)

type MockEventStore struct {
	AccessEvents []auth.SessionAccessEvent
}

var _ EventStore = &MockEventStore{}

func NewMockEventStore() *MockEventStore {
	return &MockEventStore{}
}

func (s *MockEventStore) AppendAccessEvent(_ *auth.Session, e *auth.SessionAccessEvent) error {
	s.AccessEvents = append(s.AccessEvents, *e)
	return nil
}

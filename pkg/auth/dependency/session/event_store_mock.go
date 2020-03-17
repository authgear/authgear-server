package session

import "github.com/skygeario/skygear-server/pkg/core/authn"

type MockEventStore struct {
	AccessEvents []authn.AccessEvent
}

var _ EventStore = &MockEventStore{}

func NewMockEventStore() *MockEventStore {
	return &MockEventStore{}
}

func (s *MockEventStore) AppendAccessEvent(_ *Session, e *authn.AccessEvent) error {
	s.AccessEvents = append(s.AccessEvents, *e)
	return nil
}

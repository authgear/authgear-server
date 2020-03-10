package session

type MockEventStore struct {
	AccessEvents []AccessEvent
}

var _ EventStore = &MockEventStore{}

func NewMockEventStore() *MockEventStore {
	return &MockEventStore{}
}

func (s *MockEventStore) AppendAccessEvent(_ *Session, e *AccessEvent) error {
	s.AccessEvents = append(s.AccessEvents, *e)
	return nil
}

package session

type EventStore interface {
	// AppendAccessEvent appends an access event to the session event stream
	AppendAccessEvent(s *Session, e *AccessEvent) error
}

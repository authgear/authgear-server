package auth

type AccessEventStore interface {
	// AppendAccessEvent appends an access event to the session event stream
	AppendAccessEvent(s AuthSession, e *AccessEvent) error
	// ResetEventStream resets a session event stream
	ResetEventStream(s AuthSession) error
}

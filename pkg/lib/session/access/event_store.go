package access

import (
	"time"
)

type EventStore interface {
	// AppendEvent appends an access event to the session event stream
	AppendEvent(sessionID string, expiry time.Time, e *Event) error
	// ResetEventStream resets a session event stream
	ResetEventStream(sessionID string) error
}

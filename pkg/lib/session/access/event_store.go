package access

import (
	"context"
	"time"
)

type EventStore interface {
	// AppendEvent appends an access event to the session event stream
	AppendEvent(ctx context.Context, sessionID string, expiry time.Time, e *Event) error
	// ResetEventStream resets a session event stream
	ResetEventStream(ctx context.Context, sessionID string) error
}

package session

import "github.com/skygeario/skygear-server/pkg/core/auth"

type EventStore interface {
	// AppendAccessEvent appends an access event to the session event stream
	AppendAccessEvent(s *auth.Session, e *auth.SessionAccessEvent) error
}

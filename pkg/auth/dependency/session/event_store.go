package session

import "github.com/skygeario/skygear-server/pkg/core/authn"

type EventStore interface {
	// AppendAccessEvent appends an access event to the session event stream
	AppendAccessEvent(s *Session, e *authn.AccessEvent) error
}

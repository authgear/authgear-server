package authn

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/session"
)

// SessionProvider manipulates authentication sessions.
type SessionProvider interface {
	// BeginSession creates a new authentication session.
	BeginSession(userID string, prin principal.Principal, reason session.CreateReason) (*Session, error)

	// Complete creates a session from a completed authentication session and return an authentication result.
	CompleteSession(s *Session) (Result, error)

	// MakeResult loads related data for an existing session to create an authentication result.
	MakeResult(s *session.Session) (Result, error)

	// Resolve resolves token to authentication session.
	ResolveSession(token string) (*Session, error)
}

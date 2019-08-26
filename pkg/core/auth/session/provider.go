package session

import "github.com/skygeario/skygear-server/pkg/core/auth"

type Provider interface {
	// Create creates a session for principal
	Create(userID string, principalID string) (*auth.Session, error)
	// GetByToken gets the session identified by the token
	GetByToken(token string, kind auth.SessionTokenKind) (*auth.Session, error)
	// Access updates the session info when it is being accessed by user
	Access(*auth.Session) error
	// Invalidate invalidates session with the ID
	Invalidate(id string) error

	// Refresh re-generates the access token of the session
	Refresh(*auth.Session) error
}

package session

import "github.com/skygeario/skygear-server/pkg/core/authn"

type Provider interface {
	// Make makes a session from authn attributes
	MakeSession(attrs *authn.Attrs) (session *IDPSession, token string)
	// Create creates a session
	Create(session *IDPSession) error
	// GetByToken gets the session identified by the token
	GetByToken(token string) (*IDPSession, error)
	// Get gets the session identified by the ID
	Get(id string) (*IDPSession, error)
	// Update updates the session attributes.
	Update(session *IDPSession) error
}

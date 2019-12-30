package session

import "github.com/skygeario/skygear-server/pkg/core/auth"

type Provider interface {
	// Create creates a session from AuthnSession
	Create(authnSess *auth.AuthnSession, beforeCreate func(*auth.Session) error) (*auth.Session, auth.SessionTokens, error)
	// GetByToken gets the session identified by the token
	GetByToken(token string, kind auth.SessionTokenKind) (*auth.Session, error)
	// Get gets the session identified by the ID
	Get(id string) (*auth.Session, error)
	// Access updates the session info when it is being accessed by user
	Access(*auth.Session) error
	// Invalidate invalidates session with the ID
	Invalidate(*auth.Session) error
	// InvalidateBatch invalidates sessions
	InvalidateBatch([]*auth.Session) error
	// InvalidateAll invalidates all sessions of the user, except specified session
	InvalidateAll(userID string, sessionID string) error
	// List lists the sessions belonging to the user, in ascending creation time order
	List(userID string) ([]*auth.Session, error)

	// Refresh re-generates the access token of the session
	Refresh(*auth.Session) (accessToken string, err error)

	// UpdateMFA updates the session MFA.
	UpdateMFA(sess *auth.Session, opts auth.AuthnSessionStepMFAOptions) error
	// UpdatePrincipal updates the principal ID of the session
	UpdatePrincipal(sess *auth.Session, principalID string) error
}

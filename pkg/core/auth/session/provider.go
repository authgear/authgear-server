package session

type Provider interface {
	// Create creates a session for principal
	Create(userID string, principalID string) (*Session, error)
	// GetByToken gets the session identified by the token
	GetByToken(token string, kind TokenKind) (*Session, error)
	// Access updates the session info when it is being accessed by user
	Access(*Session) error
	// Invalidate invalidates session with the ID
	Invalidate(id string) error
}

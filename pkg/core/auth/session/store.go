package session

import "fmt"

var ErrSessionNotFound = fmt.Errorf("session is not found")

type Store interface {
	// Create creates a session in the store. It must not allow overwriting existing sessions.
	Create(s *Session) error
	// Update updates a session in the store. It must return `ErrSessionNotFound` when the session does not exist.
	Update(s *Session) error
	// Get returns the session with id in the store. It must return `ErrSessionNotFound` when the session does not exist.
	Get(id string) (*Session, error)
	// Delete deletes the session with id in the store. It must treat deleting non-existent session as successful.
	Delete(id string) error
}

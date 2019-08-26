package session

import (
	"fmt"
	"time"

	"github.com/skygeario/skygear-server/pkg/core/auth"
)

var ErrSessionNotFound = fmt.Errorf("session is not found")

type Store interface {
	// Create creates a session in the store. It must not allow overwriting existing sessions.
	Create(s *auth.Session, ttl time.Duration) error
	// Update updates a session in the store. It must return `ErrSessionNotFound` when the session does not exist.
	Update(s *auth.Session, ttl time.Duration) error
	// Get returns the session with id in the store. It must return `ErrSessionNotFound` when the session does not exist.
	Get(id string) (*auth.Session, error)
	// Delete deletes the session with id in the store. It must treat deleting non-existent session as successful.
	Delete(id string) error
}

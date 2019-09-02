package session

import (
	"fmt"
	"time"

	"github.com/skygeario/skygear-server/pkg/core/auth"
)

var ErrSessionNotFound = fmt.Errorf("session is not found")

// Store represents the backing store for user sessions.
// Note that the returned sessions may not be valid (e.g. can be expired)
type Store interface {
	// Create creates a session in the store. It must not allow overwriting existing sessions.
	Create(s *auth.Session, expireAt time.Time) error
	// Update updates a session in the store. It must return `ErrSessionNotFound` when the session does not exist.
	Update(s *auth.Session, expireAt time.Time) error
	// Get returns the session with id in the store. It must return `ErrSessionNotFound` when the session does not exist.
	Get(id string) (*auth.Session, error)
	// Delete deletes the session with id in the store. It must treat deleting non-existent session as successful.
	Delete(*auth.Session) error
	// DeleteAll deletes all sessions of the user in the store, excluding specified session.
	DeleteAll(userID string, sessionID string) error
	// List lists the sessions belonging to the user, in ascending creation time order
	List(userID string) ([]*auth.Session, error)
}

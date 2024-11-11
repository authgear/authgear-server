package idpsession

import (
	"context"
	"time"
)

//go:generate mockgen -source=store.go -destination=store_mock_test.go -package idpsession

// Store represents the backing store for IdP sessions.
// Note that the returned sessions may not be valid (e.g. can be expired)
type Store interface {
	// Create creates a session in the Store. It must not allow overwriting existing sessions.
	Create(ctx context.Context, s *IDPSession, expireAt time.Time) error
	// Update updates a session in the Store. It must return `ErrSessionNotFound` when the session does not exist.
	Update(ctx context.Context, s *IDPSession, expireAt time.Time) error
	// Get returns the session with id in the Store. It must return `ErrSessionNotFound` when the session does not exist.
	Get(ctx context.Context, id string) (*IDPSession, error)
	// Delete deletes the session with id in the Store. It must treat deleting non-existent session as successful.
	Delete(ctx context.Context, s *IDPSession) error
	// List lists the sessions belonging to the user, in ascending creation time order
	List(ctx context.Context, userID string) ([]*IDPSession, error)
	// CleanUpForDeletingUserID cleans up for a deleting user ID.
	CleanUpForDeletingUserID(ctx context.Context, userID string) error
}

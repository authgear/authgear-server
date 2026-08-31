package nonblocking

import (
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
)

// EventPayloadInitialAccessToken describes an initial access token in an
// audit event payload. Shared by oauth.client.registered and
// oauth.client.registration.failed, deliberately -- unlike the per-event
// Client payload structs, which are kept independent on purpose, this one
// must never diverge: the point is that a token presents identically in
// both records, so an auditor correlating "which clients did leaked token
// X register, and when did it stop working" reads one shape throughout.
//
// Never carries the token value or a hash of it (a hash is still a
// guessing oracle). ID is the row uuid the Admin API already exposes.
type EventPayloadInitialAccessToken struct {
	ID        string                            `json:"id"`
	Type      model.OAuthInitialAccessTokenType `json:"type"`
	CreatedAt time.Time                         `json:"created_at"`
	ExpiresAt time.Time                         `json:"expires_at"`
}

// NewEventPayloadInitialAccessToken maps nil to nil, which is what both
// callers want: open registration presents no token, and an unknown or
// absent one has no row to describe. Combined with `omitempty` on the
// field, that drops the key entirely rather than emitting an object of
// zero values.
func NewEventPayloadInitialAccessToken(iat *model.OAuthInitialAccessToken) *EventPayloadInitialAccessToken {
	if iat == nil {
		return nil
	}
	return &EventPayloadInitialAccessToken{
		ID:        iat.ID,
		Type:      iat.Type,
		CreatedAt: iat.CreatedAt,
		ExpiresAt: iat.ExpiresAt,
	}
}

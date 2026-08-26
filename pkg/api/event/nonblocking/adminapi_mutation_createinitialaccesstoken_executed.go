package nonblocking

import (
	"time"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AdminAPIMutationCreateInitialAccessTokenExecuted event.Type = "admin_api.mutation.create_initial_access_token.executed" // #nosec G101
)

// AdminAPIMutationCreateInitialAccessTokenExecutedEventPayload deliberately
// carries no token material. The plaintext IAT returned by
// createInitialAccessToken is the one-and-only copy (docs/specs/dcr.md — IAT
// storage); putting it in an audit row would make a 90-day-retained,
// admin-readable second copy of a credential the design says is
// unrecoverable. Only the hashed form exists in
// _auth_oauth_initial_access_token, and only the id/type/expiry go here.
type AdminAPIMutationCreateInitialAccessTokenExecutedEventPayload struct {
	InitialAccessTokenID string                            `json:"initial_access_token_id"`
	Type                 model.OAuthInitialAccessTokenType `json:"type"`
	ExpiresAt            time.Time                         `json:"expires_at"`
}

func (e *AdminAPIMutationCreateInitialAccessTokenExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIMutationCreateInitialAccessTokenExecuted
}

func (e *AdminAPIMutationCreateInitialAccessTokenExecutedEventPayload) UserID() string {
	return ""
}

func (e *AdminAPIMutationCreateInitialAccessTokenExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIMutationCreateInitialAccessTokenExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPIMutationCreateInitialAccessTokenExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIMutationCreateInitialAccessTokenExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPIMutationCreateInitialAccessTokenExecutedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *AdminAPIMutationCreateInitialAccessTokenExecutedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &AdminAPIMutationCreateInitialAccessTokenExecutedEventPayload{}

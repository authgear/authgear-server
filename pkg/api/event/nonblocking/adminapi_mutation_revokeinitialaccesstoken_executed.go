package nonblocking

import (
	"time"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AdminAPIMutationRevokeInitialAccessTokenExecuted event.Type = "admin_api.mutation.revoke_initial_access_token.executed" // #nosec G101
)

// AdminAPIMutationRevokeInitialAccessTokenExecutedEventPayload carries the
// revoked token's id/type/expiry, read from the row before it was deleted.
// Type is recorded because revoking a first-party IAT is a materially
// different administrative act from revoking a third-party one.
type AdminAPIMutationRevokeInitialAccessTokenExecutedEventPayload struct {
	InitialAccessTokenID string                            `json:"initial_access_token_id"`
	Type                 model.OAuthInitialAccessTokenType `json:"type"`
	ExpiresAt            time.Time                         `json:"expires_at"`
}

func (e *AdminAPIMutationRevokeInitialAccessTokenExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIMutationRevokeInitialAccessTokenExecuted
}

func (e *AdminAPIMutationRevokeInitialAccessTokenExecutedEventPayload) UserID() string {
	return ""
}

func (e *AdminAPIMutationRevokeInitialAccessTokenExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIMutationRevokeInitialAccessTokenExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPIMutationRevokeInitialAccessTokenExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIMutationRevokeInitialAccessTokenExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPIMutationRevokeInitialAccessTokenExecutedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *AdminAPIMutationRevokeInitialAccessTokenExecutedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &AdminAPIMutationRevokeInitialAccessTokenExecutedEventPayload{}

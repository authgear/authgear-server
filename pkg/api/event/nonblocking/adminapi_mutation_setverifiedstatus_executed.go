package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
)

const (
	AdminAPIMutationSetVerifiedStatusExecuted event.Type = "admin_api.mutation.set_verified_status.executed"
)

type AdminAPIMutationSetVerifiedStatusExecutedEventPayload struct {
	ClaimName  string `json:"claim_name"`
	ClaimValue string `json:"claim_value"`
	IsVerified bool   `json:"is_verified"`
}

func (e *AdminAPIMutationSetVerifiedStatusExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIMutationSetVerifiedStatusExecuted
}

func (e *AdminAPIMutationSetVerifiedStatusExecutedEventPayload) UserID() string {
	return ""
}

func (e *AdminAPIMutationSetVerifiedStatusExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIMutationSetVerifiedStatusExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPIMutationSetVerifiedStatusExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIMutationSetVerifiedStatusExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPIMutationSetVerifiedStatusExecutedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *AdminAPIMutationSetVerifiedStatusExecutedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &AdminAPIMutationSetVerifiedStatusExecutedEventPayload{}

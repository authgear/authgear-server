package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
)

const (
	AdminAPISetVerifiedStatusExecuted event.Type = "admin_api.set_verified_status.executed"
)

type AdminAPISetVerifiedStatusExecutedEventPayload struct {
	ClaimName  string `json:"claim_name"`
	ClaimValue string `json:"claim_value"`
	IsVerified bool   `json:"is_verified"`
}

func (e *AdminAPISetVerifiedStatusExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPISetVerifiedStatusExecuted
}

func (e *AdminAPISetVerifiedStatusExecutedEventPayload) UserID() string {
	return ""
}

func (e *AdminAPISetVerifiedStatusExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPISetVerifiedStatusExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPISetVerifiedStatusExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPISetVerifiedStatusExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPISetVerifiedStatusExecutedEventPayload) ReindexUserNeeded() bool {
	return false
}

func (e *AdminAPISetVerifiedStatusExecutedEventPayload) IsUserDeleted() bool {
	return false
}

var _ event.NonBlockingPayload = &AdminAPISetVerifiedStatusExecutedEventPayload{}

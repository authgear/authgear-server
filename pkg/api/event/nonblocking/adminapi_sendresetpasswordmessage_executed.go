package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
)

const (
	AdminAPISendResetPasswordMessageExecuted event.Type = "admin_api.send_reset_password_message.executed"
)

type AdminAPISendResetPasswordMessageExecutedEventPayload struct {
	LoginID string `json:"login_id"`
}

func (e *AdminAPISendResetPasswordMessageExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPISendResetPasswordMessageExecuted
}

func (e *AdminAPISendResetPasswordMessageExecutedEventPayload) UserID() string {
	return ""
}

func (e *AdminAPISendResetPasswordMessageExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPISendResetPasswordMessageExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPISendResetPasswordMessageExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPISendResetPasswordMessageExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPISendResetPasswordMessageExecutedEventPayload) ReindexUserNeeded() bool {
	return false
}

func (e *AdminAPISendResetPasswordMessageExecutedEventPayload) IsUserDeleted() bool {
	return false
}

var _ event.NonBlockingPayload = &AdminAPISendResetPasswordMessageExecutedEventPayload{}

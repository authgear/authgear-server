package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
)

const (
	AdminAPIMutationSendResetPasswordMessageExecuted event.Type = "admin_api.mutation.send_reset_password_message.executed" // nolint:gosec
)

type AdminAPIMutationSendResetPasswordMessageExecutedEventPayload struct {
	LoginID string `json:"login_id"`
}

func (e *AdminAPIMutationSendResetPasswordMessageExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIMutationSendResetPasswordMessageExecuted
}

func (e *AdminAPIMutationSendResetPasswordMessageExecutedEventPayload) UserID() string {
	return ""
}

func (e *AdminAPIMutationSendResetPasswordMessageExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIMutationSendResetPasswordMessageExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPIMutationSendResetPasswordMessageExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIMutationSendResetPasswordMessageExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPIMutationSendResetPasswordMessageExecutedEventPayload) RequireReindexUserIDs() []string {
	return []string{}
}

func (e *AdminAPIMutationSendResetPasswordMessageExecutedEventPayload) DeletedUserIDs() []string {
	return []string{}
}

var _ event.NonBlockingPayload = &AdminAPIMutationSendResetPasswordMessageExecutedEventPayload{}

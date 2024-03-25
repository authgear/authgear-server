package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AdminAPIMutationResetPasswordExecuted event.Type = "admin_api.mutation.reset_password.executed" // nolint:gosec
)

type AdminAPIMutationResetPasswordExecutedEventPayload struct {
	UserRef   model.UserRef `json:"-" resolve:"user"`
	UserModel model.User    `json:"user"`
}

func (e *AdminAPIMutationResetPasswordExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIMutationResetPasswordExecuted
}

func (e *AdminAPIMutationResetPasswordExecutedEventPayload) UserID() string {
	return e.UserModel.ID
}

func (e *AdminAPIMutationResetPasswordExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIMutationResetPasswordExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPIMutationResetPasswordExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIMutationResetPasswordExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPIMutationResetPasswordExecutedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *AdminAPIMutationResetPasswordExecutedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &AdminAPIMutationResetPasswordExecutedEventPayload{}

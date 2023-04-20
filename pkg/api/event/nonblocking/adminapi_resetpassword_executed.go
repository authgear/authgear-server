package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AdminAPIResetPasswordExecuted event.Type = "admin_api.reset_password.executed"
)

type AdminAPIResetPasswordExecutedEventPayload struct {
	UserRef   model.UserRef `json:"-" resolve:"user"`
	UserModel model.User    `json:"user"`
}

func (e *AdminAPIResetPasswordExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIResetPasswordExecuted
}

func (e *AdminAPIResetPasswordExecutedEventPayload) UserID() string {
	return e.UserModel.ID
}

func (e *AdminAPIResetPasswordExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIResetPasswordExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPIResetPasswordExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIResetPasswordExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPIResetPasswordExecutedEventPayload) ReindexUserNeeded() bool {
	return false
}

func (e *AdminAPIResetPasswordExecutedEventPayload) IsUserDeleted() bool {
	return false
}

var _ event.NonBlockingPayload = &AdminAPIResetPasswordExecutedEventPayload{}

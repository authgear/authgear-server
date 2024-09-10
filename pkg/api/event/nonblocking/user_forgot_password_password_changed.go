package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	UserForgotPasswordPasswordChanged event.Type = "user.forgot_password.password_changed"
)

type UserForgotPasswordPasswordChangedEventPayload struct {
	UserRef   model.UserRef   `json:"-" resolve:"user"`
	UserModel model.User      `json:"user"`
	Sessions  []model.Session `json:"sessions"`
}

func (e *UserForgotPasswordPasswordChangedEventPayload) NonBlockingEventType() event.Type {
	return UserForgotPasswordPasswordChanged
}

func (e *UserForgotPasswordPasswordChangedEventPayload) UserID() string {
	return e.UserRef.ID
}

func (e *UserForgotPasswordPasswordChangedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeUser
}

func (e *UserForgotPasswordPasswordChangedEventPayload) FillContext(ctx *event.Context) {
}

func (e *UserForgotPasswordPasswordChangedEventPayload) ForHook() bool {
	return false
}

func (e *UserForgotPasswordPasswordChangedEventPayload) ForAudit() bool {
	return true
}

func (e *UserForgotPasswordPasswordChangedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *UserForgotPasswordPasswordChangedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &UserForgotPasswordPasswordChangedEventPayload{}

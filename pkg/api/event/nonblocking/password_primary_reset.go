package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	PasswordPrimaryReset event.Type = "password.primary.reset"
)

type PasswordPrimaryResetEventPayload struct {
	UserRef   model.UserRef `json:"-" resolve:"user"`
	UserModel model.User    `json:"user"`
}

func (e *PasswordPrimaryResetEventPayload) NonBlockingEventType() event.Type {
	return PasswordPrimaryReset
}

func (e *PasswordPrimaryResetEventPayload) UserID() string {
	return e.UserRef.ID
}

func (e *PasswordPrimaryResetEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeUser
}

func (e *PasswordPrimaryResetEventPayload) FillContext(ctx *event.Context) {
}

func (e *PasswordPrimaryResetEventPayload) ForHook() bool {
	return false
}

func (e *PasswordPrimaryResetEventPayload) ForAudit() bool {
	return true
}

func (e *PasswordPrimaryResetEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *PasswordPrimaryResetEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &PasswordPrimaryResetEventPayload{}

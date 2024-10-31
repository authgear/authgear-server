package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	PasswordPrimaryChanged event.Type = "password.primary.changed"
)

type PasswordPrimaryChangedEventPayload struct {
	UserRef   model.UserRef `json:"-" resolve:"user"`
	UserModel model.User    `json:"user"`
}

func (e *PasswordPrimaryChangedEventPayload) NonBlockingEventType() event.Type {
	return PasswordPrimaryChanged
}

func (e *PasswordPrimaryChangedEventPayload) UserID() string {
	return e.UserRef.ID
}

func (e *PasswordPrimaryChangedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeUser
}

func (e *PasswordPrimaryChangedEventPayload) FillContext(ctx *event.Context) {
}

func (e *PasswordPrimaryChangedEventPayload) ForHook() bool {
	return false
}

func (e *PasswordPrimaryChangedEventPayload) ForAudit() bool {
	return true
}

func (e *PasswordPrimaryChangedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *PasswordPrimaryChangedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &PasswordPrimaryChangedEventPayload{}

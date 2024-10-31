package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	PasswordSecondaryChanged event.Type = "password.secondary.changed"
)

type PasswordSecondaryChangedEventPayload struct {
	UserRef   model.UserRef `json:"-" resolve:"user"`
	UserModel model.User    `json:"user"`
}

func (e *PasswordSecondaryChangedEventPayload) NonBlockingEventType() event.Type {
	return PasswordSecondaryChanged
}

func (e *PasswordSecondaryChangedEventPayload) UserID() string {
	return e.UserRef.ID
}

func (e *PasswordSecondaryChangedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeUser
}

func (e *PasswordSecondaryChangedEventPayload) FillContext(ctx *event.Context) {
}

func (e *PasswordSecondaryChangedEventPayload) ForHook() bool {
	return false
}

func (e *PasswordSecondaryChangedEventPayload) ForAudit() bool {
	return true
}

func (e *PasswordSecondaryChangedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *PasswordSecondaryChangedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &PasswordSecondaryChangedEventPayload{}

package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	UserForceUpdatePasswordChanged event.Type = "user.force_update.password_changed"
)

type UserForceUpdatePasswordChangedEventPayload struct {
	UserRef   model.UserRef   `json:"-" resolve:"user"`
	UserModel model.User      `json:"user"`
	Sessions  []model.Session `json:"sessions"`
}

func (e *UserForceUpdatePasswordChangedEventPayload) NonBlockingEventType() event.Type {
	return UserForceUpdatePasswordChanged
}

func (e *UserForceUpdatePasswordChangedEventPayload) UserID() string {
	return e.UserRef.ID
}

func (e *UserForceUpdatePasswordChangedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeUser
}

func (e *UserForceUpdatePasswordChangedEventPayload) FillContext(ctx *event.Context) {
}

func (e *UserForceUpdatePasswordChangedEventPayload) ForHook() bool {
	return false
}

func (e *UserForceUpdatePasswordChangedEventPayload) ForAudit() bool {
	return true
}

func (e *UserForceUpdatePasswordChangedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *UserForceUpdatePasswordChangedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &UserForceUpdatePasswordChangedEventPayload{}

package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	UserSettingsPasswordChanged event.Type = "user.settings.password_changed"
)

type UserSettingsPasswordChangedEventPayload struct {
	UserRef   model.UserRef   `json:"-" resolve:"user"`
	UserModel model.User      `json:"user"`
	Sessions  []model.Session `json:"sessions"`
}

func (e *UserSettingsPasswordChangedEventPayload) NonBlockingEventType() event.Type {
	return UserSettingsPasswordChanged
}

func (e *UserSettingsPasswordChangedEventPayload) UserID() string {
	return e.UserRef.ID
}

func (e *UserSettingsPasswordChangedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeUser
}

func (e *UserSettingsPasswordChangedEventPayload) FillContext(ctx *event.Context) {
}

func (e *UserSettingsPasswordChangedEventPayload) ForHook() bool {
	return false
}

func (e *UserSettingsPasswordChangedEventPayload) ForAudit() bool {
	return true
}

func (e *UserSettingsPasswordChangedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *UserSettingsPasswordChangedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &UserSettingsPasswordChangedEventPayload{}

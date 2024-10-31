package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	UserSettingsSecondaryPasswordChanged event.Type = "user.settings.secondary_password_changed"
)

type UserSettingsSecondaryPasswordChangedEventPayload struct {
	UserRef   model.UserRef `json:"-" resolve:"user"`
	UserModel model.User    `json:"user"`
}

func (e *UserSettingsSecondaryPasswordChangedEventPayload) NonBlockingEventType() event.Type {
	return UserSettingsSecondaryPasswordChanged
}

func (e *UserSettingsSecondaryPasswordChangedEventPayload) UserID() string {
	return e.UserRef.ID
}

func (e *UserSettingsSecondaryPasswordChangedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeUser
}

func (e *UserSettingsSecondaryPasswordChangedEventPayload) FillContext(ctx *event.Context) {
}

func (e *UserSettingsSecondaryPasswordChangedEventPayload) ForHook() bool {
	return false
}

func (e *UserSettingsSecondaryPasswordChangedEventPayload) ForAudit() bool {
	return true
}

func (e *UserSettingsSecondaryPasswordChangedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *UserSettingsSecondaryPasswordChangedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &UserSettingsSecondaryPasswordChangedEventPayload{}

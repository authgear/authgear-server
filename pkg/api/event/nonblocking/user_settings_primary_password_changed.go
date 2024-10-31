package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	UserSettingsPrimaryPasswordChanged event.Type = "user.settings.primary_password_changed"
)

type UserSettingsPrimaryPasswordChangedEventPayload struct {
	UserRef   model.UserRef `json:"-" resolve:"user"`
	UserModel model.User    `json:"user"`
}

func (e *UserSettingsPrimaryPasswordChangedEventPayload) NonBlockingEventType() event.Type {
	return UserSettingsPrimaryPasswordChanged
}

func (e *UserSettingsPrimaryPasswordChangedEventPayload) UserID() string {
	return e.UserRef.ID
}

func (e *UserSettingsPrimaryPasswordChangedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeUser
}

func (e *UserSettingsPrimaryPasswordChangedEventPayload) FillContext(ctx *event.Context) {
}

func (e *UserSettingsPrimaryPasswordChangedEventPayload) ForHook() bool {
	return false
}

func (e *UserSettingsPrimaryPasswordChangedEventPayload) ForAudit() bool {
	return true
}

func (e *UserSettingsPrimaryPasswordChangedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *UserSettingsPrimaryPasswordChangedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &UserSettingsPrimaryPasswordChangedEventPayload{}

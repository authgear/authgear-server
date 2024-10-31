package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	// nolint:gosec // false-positive of lint
	PasswordPrimaryForceChanged event.Type = "password.primary.force_changed"
)

type PasswordPrimaryForceChangedEventPayload struct {
	UserRef   model.UserRef `json:"-" resolve:"user"`
	UserModel model.User    `json:"user"`
	Reason    string        `json:"reason,omitempty"`
}

func (e *PasswordPrimaryForceChangedEventPayload) NonBlockingEventType() event.Type {
	return PasswordPrimaryForceChanged
}

func (e *PasswordPrimaryForceChangedEventPayload) UserID() string {
	return e.UserRef.ID
}

func (e *PasswordPrimaryForceChangedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeUser
}

func (e *PasswordPrimaryForceChangedEventPayload) FillContext(ctx *event.Context) {
}

func (e *PasswordPrimaryForceChangedEventPayload) ForHook() bool {
	return false
}

func (e *PasswordPrimaryForceChangedEventPayload) ForAudit() bool {
	return true
}

func (e *PasswordPrimaryForceChangedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *PasswordPrimaryForceChangedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &PasswordPrimaryForceChangedEventPayload{}

package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	// nolint:gosec // false-positive of lint
	PasswordSecondaryForceChanged event.Type = "password.secondary.force_changed"
)

type PasswordSecondaryForceChangedEventPayload struct {
	UserRef   model.UserRef `json:"-" resolve:"user"`
	UserModel model.User    `json:"user"`
	Reason    string        `json:"reason,omitempty"`
}

func (e *PasswordSecondaryForceChangedEventPayload) NonBlockingEventType() event.Type {
	return PasswordSecondaryForceChanged
}

func (e *PasswordSecondaryForceChangedEventPayload) UserID() string {
	return e.UserRef.ID
}

func (e *PasswordSecondaryForceChangedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeUser
}

func (e *PasswordSecondaryForceChangedEventPayload) FillContext(ctx *event.Context) {
}

func (e *PasswordSecondaryForceChangedEventPayload) ForHook() bool {
	return false
}

func (e *PasswordSecondaryForceChangedEventPayload) ForAudit() bool {
	return true
}

func (e *PasswordSecondaryForceChangedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *PasswordSecondaryForceChangedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &PasswordSecondaryForceChangedEventPayload{}

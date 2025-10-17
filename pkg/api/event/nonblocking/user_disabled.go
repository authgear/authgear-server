package nonblocking

import (
	"time"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	UserDisabled event.Type = "user.disabled"
)

type UserDisabledEventPayload struct {
	UserRef                  model.UserRef `json:"-" resolve:"user"`
	UserModel                model.User    `json:"user"`
	TemporarilyDisabledFrom  *time.Time    `json:"temporarily_disabled_from,omitempty"`
	TemporarilyDisabledUntil *time.Time    `json:"temporarily_disabled_until,omitempty"`
}

func (e *UserDisabledEventPayload) NonBlockingEventType() event.Type {
	return UserDisabled
}

func (e *UserDisabledEventPayload) UserID() string {
	return e.UserRef.ID
}

func (e *UserDisabledEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *UserDisabledEventPayload) FillContext(ctx *event.Context) {
}

func (e *UserDisabledEventPayload) ForHook() bool {
	return true
}

func (e *UserDisabledEventPayload) ForAudit() bool {
	return true
}

func (e *UserDisabledEventPayload) RequireReindexUserIDs() []string {
	return []string{e.UserID()}
}

func (e *UserDisabledEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &UserDisabledEventPayload{}

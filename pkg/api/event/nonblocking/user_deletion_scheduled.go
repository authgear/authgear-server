package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	UserDeletionScheduled event.Type = "user.deletion_scheduled"
)

type UserDeletionScheduledEventPayload struct {
	UserRef   model.UserRef `json:"-" resolve:"user"`
	UserModel model.User    `json:"user"`
	AdminAPI  bool          `json:"-"`
}

func (e *UserDeletionScheduledEventPayload) NonBlockingEventType() event.Type {
	return UserDeletionScheduled
}

func (e *UserDeletionScheduledEventPayload) UserID() string {
	return e.UserRef.ID
}

func (e *UserDeletionScheduledEventPayload) GetTriggeredBy() event.TriggeredByType {
	if e.AdminAPI {
		return event.TriggeredByTypeAdminAPI
	}
	return event.TriggeredByTypeUser
}

func (e *UserDeletionScheduledEventPayload) FillContext(ctx *event.Context) {}

func (e *UserDeletionScheduledEventPayload) ForWebHook() bool {
	return true
}

func (e *UserDeletionScheduledEventPayload) ForAudit() bool {
	return true
}

func (e *UserDeletionScheduledEventPayload) ReindexUserNeeded() bool {
	return true
}

func (e *UserDeletionScheduledEventPayload) IsUserDeleted() bool {
	return false
}

var _ event.NonBlockingPayload = &UserDeletionScheduledEventPayload{}

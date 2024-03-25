package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	UserDeletionUnscheduled event.Type = "user.deletion_unscheduled"
)

type UserDeletionUnscheduledEventPayload struct {
	UserRef   model.UserRef `json:"-" resolve:"user"`
	UserModel model.User    `json:"user"`
	AdminAPI  bool          `json:"-"`
}

func (e *UserDeletionUnscheduledEventPayload) NonBlockingEventType() event.Type {
	return UserDeletionUnscheduled
}

func (e *UserDeletionUnscheduledEventPayload) UserID() string {
	return e.UserRef.ID
}

func (e *UserDeletionUnscheduledEventPayload) GetTriggeredBy() event.TriggeredByType {
	if e.AdminAPI {
		return event.TriggeredByTypeAdminAPI
	}
	return event.TriggeredByTypeUser
}

func (e *UserDeletionUnscheduledEventPayload) FillContext(ctx *event.Context) {}

func (e *UserDeletionUnscheduledEventPayload) ForHook() bool {
	return true
}

func (e *UserDeletionUnscheduledEventPayload) ForAudit() bool {
	return true
}

func (e *UserDeletionUnscheduledEventPayload) RequireReindexUserIDs() []string {
	return []string{e.UserID()}
}

func (e *UserDeletionUnscheduledEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &UserDeletionUnscheduledEventPayload{}

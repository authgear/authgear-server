package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	UserAnonymizationUnscheduled event.Type = "user.anonymization_unscheduled"
)

type UserAnonymizationUnscheduledEventPayload struct {
	UserRef   model.UserRef `json:"-" resolve:"user"`
	UserModel model.User    `json:"user"`
	AdminAPI  bool          `json:"-"`
}

func (e *UserAnonymizationUnscheduledEventPayload) NonBlockingEventType() event.Type {
	return UserAnonymizationUnscheduled
}

func (e *UserAnonymizationUnscheduledEventPayload) UserID() string {
	return e.UserRef.ID
}

func (e *UserAnonymizationUnscheduledEventPayload) GetTriggeredBy() event.TriggeredByType {
	if e.AdminAPI {
		return event.TriggeredByTypeAdminAPI
	}
	return event.TriggeredByTypeUser
}

func (e *UserAnonymizationUnscheduledEventPayload) FillContext(ctx *event.Context) {}

func (e *UserAnonymizationUnscheduledEventPayload) ForHook() bool {
	return true
}

func (e *UserAnonymizationUnscheduledEventPayload) ForAudit() bool {
	return true
}

func (e *UserAnonymizationUnscheduledEventPayload) RequireReindexUserIDs() []string {
	return []string{e.UserID()}
}

func (e *UserAnonymizationUnscheduledEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &UserAnonymizationUnscheduledEventPayload{}

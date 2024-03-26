package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	UserAnonymizationScheduled event.Type = "user.anonymization_scheduled"
)

type UserAnonymizationScheduledEventPayload struct {
	UserRef   model.UserRef `json:"-" resolve:"user"`
	UserModel model.User    `json:"user"`
	AdminAPI  bool          `json:"-"`
}

func (e *UserAnonymizationScheduledEventPayload) NonBlockingEventType() event.Type {
	return UserAnonymizationScheduled
}

func (e *UserAnonymizationScheduledEventPayload) UserID() string {
	return e.UserRef.ID
}

func (e *UserAnonymizationScheduledEventPayload) GetTriggeredBy() event.TriggeredByType {
	if e.AdminAPI {
		return event.TriggeredByTypeAdminAPI
	}
	return event.TriggeredByTypeUser
}

func (e *UserAnonymizationScheduledEventPayload) FillContext(ctx *event.Context) {}

func (e *UserAnonymizationScheduledEventPayload) ForHook() bool {
	return true
}

func (e *UserAnonymizationScheduledEventPayload) ForAudit() bool {
	return true
}

func (e *UserAnonymizationScheduledEventPayload) RequireReindexUserIDs() []string {
	return []string{e.UserID()}
}

func (e *UserAnonymizationScheduledEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &UserAnonymizationScheduledEventPayload{}

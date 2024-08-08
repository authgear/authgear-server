package nonblocking

import (
	"time"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	UserSetMFAGracePeriod event.Type = "user.anonymization_unscheduled"
)

type UserSetMFAGracePeriodEventPayload struct {
	UserRef   model.UserRef `json:"-" resolve:"user"`
	UserModel model.User    `json:"user"`
	AdminAPI  bool          `json:"-"`
	EndAt     *time.Time    `json:"end_at"`
}

func (e *UserSetMFAGracePeriodEventPayload) NonBlockingEventType() event.Type {
	return UserSetMFAGracePeriod
}

func (e *UserSetMFAGracePeriodEventPayload) UserID() string {
	return e.UserRef.ID
}

func (e *UserSetMFAGracePeriodEventPayload) GetTriggeredBy() event.TriggeredByType {
	if e.AdminAPI {
		return event.TriggeredByTypeAdminAPI
	}
	return event.TriggeredByTypeUser
}

func (e *UserSetMFAGracePeriodEventPayload) FillContext(ctx *event.Context) {}

func (e *UserSetMFAGracePeriodEventPayload) ForHook() bool {
	return true
}

func (e *UserSetMFAGracePeriodEventPayload) ForAudit() bool {
	return true
}

func (e *UserSetMFAGracePeriodEventPayload) RequireReindexUserIDs() []string {
	return []string{e.UserID()}
}

func (e *UserSetMFAGracePeriodEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &UserSetMFAGracePeriodEventPayload{}

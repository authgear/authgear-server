package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AdminAPIUnscheduleAccountDeletionExecuted event.Type = "admin_api.unschedule_account_deletion.executed"
)

type AdminAPIUnscheduleAccountDeletionExecutedEventPayload struct {
	UserRef   model.UserRef `json:"-" resolve:"user"`
	UserModel model.User    `json:"user"`
}

func (e *AdminAPIUnscheduleAccountDeletionExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIUnscheduleAccountDeletionExecuted
}

func (e *AdminAPIUnscheduleAccountDeletionExecutedEventPayload) UserID() string {
	return e.UserModel.ID
}

func (e *AdminAPIUnscheduleAccountDeletionExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIUnscheduleAccountDeletionExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPIUnscheduleAccountDeletionExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIUnscheduleAccountDeletionExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPIUnscheduleAccountDeletionExecutedEventPayload) ReindexUserNeeded() bool {
	return false
}

func (e *AdminAPIUnscheduleAccountDeletionExecutedEventPayload) IsUserDeleted() bool {
	return false
}

var _ event.NonBlockingPayload = &AdminAPIUnscheduleAccountDeletionExecutedEventPayload{}

package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AdminAPIScheduleAccountDeletionExecuted event.Type = "admin_api.schedule_account_deletion.executed"
)

type AdminAPIScheduleAccountDeletionExecutedEventPayload struct {
	UserRef   model.UserRef `json:"-" resolve:"user"`
	UserModel model.User    `json:"user"`
}

func (e *AdminAPIScheduleAccountDeletionExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIScheduleAccountDeletionExecuted
}

func (e *AdminAPIScheduleAccountDeletionExecutedEventPayload) UserID() string {
	return e.UserModel.ID
}

func (e *AdminAPIScheduleAccountDeletionExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIScheduleAccountDeletionExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPIScheduleAccountDeletionExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIScheduleAccountDeletionExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPIScheduleAccountDeletionExecutedEventPayload) ReindexUserNeeded() bool {
	return false
}

func (e *AdminAPIScheduleAccountDeletionExecutedEventPayload) IsUserDeleted() bool {
	return false
}

var _ event.NonBlockingPayload = &AdminAPIScheduleAccountDeletionExecutedEventPayload{}

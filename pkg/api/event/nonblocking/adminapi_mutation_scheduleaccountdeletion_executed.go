package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AdminAPIMutationScheduleAccountDeletionExecuted event.Type = "admin_api.mutation.schedule_account_deletion.executed"
)

type AdminAPIMutationScheduleAccountDeletionExecutedEventPayload struct {
	UserRef   model.UserRef `json:"-" resolve:"user"`
	UserModel model.User    `json:"user"`
}

func (e *AdminAPIMutationScheduleAccountDeletionExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIMutationScheduleAccountDeletionExecuted
}

func (e *AdminAPIMutationScheduleAccountDeletionExecutedEventPayload) UserID() string {
	return e.UserModel.ID
}

func (e *AdminAPIMutationScheduleAccountDeletionExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIMutationScheduleAccountDeletionExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPIMutationScheduleAccountDeletionExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIMutationScheduleAccountDeletionExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPIMutationScheduleAccountDeletionExecutedEventPayload) RequireReindexUserIDs() []string {
	return []string{e.UserID()}
}

func (e *AdminAPIMutationScheduleAccountDeletionExecutedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &AdminAPIMutationScheduleAccountDeletionExecutedEventPayload{}

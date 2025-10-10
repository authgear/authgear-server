package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AdminAPIMutationUnscheduleAccountDeletionExecuted event.Type = "admin_api.mutation.unschedule_account_deletion.executed"
)

type AdminAPIMutationUnscheduleAccountDeletionExecutedEventPayload struct {
	UserRef   model.UserRef `json:"-" resolve:"user"`
	UserModel model.User    `json:"user"`
}

func (e *AdminAPIMutationUnscheduleAccountDeletionExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIMutationUnscheduleAccountDeletionExecuted
}

func (e *AdminAPIMutationUnscheduleAccountDeletionExecutedEventPayload) UserID() string {
	return e.UserModel.ID
}

func (e *AdminAPIMutationUnscheduleAccountDeletionExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIMutationUnscheduleAccountDeletionExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPIMutationUnscheduleAccountDeletionExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIMutationUnscheduleAccountDeletionExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPIMutationUnscheduleAccountDeletionExecutedEventPayload) RequireReindexUserIDs() []string {
	return []string{e.UserID()}
}

func (e *AdminAPIMutationUnscheduleAccountDeletionExecutedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &AdminAPIMutationUnscheduleAccountDeletionExecutedEventPayload{}

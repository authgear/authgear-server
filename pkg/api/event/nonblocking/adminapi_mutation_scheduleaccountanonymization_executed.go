package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AdminAPIMutationScheduleAccountAnonymizationExecuted event.Type = "admin_api.mutation.schedule_account_anonymization.executed"
)

type AdminAPIMutationScheduleAccountAnonymizationExecutedEventPayload struct {
	UserRef   model.UserRef `json:"-" resolve:"user"`
	UserModel model.User    `json:"user"`
}

func (e *AdminAPIMutationScheduleAccountAnonymizationExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIMutationScheduleAccountAnonymizationExecuted
}

func (e *AdminAPIMutationScheduleAccountAnonymizationExecutedEventPayload) UserID() string {
	return e.UserModel.ID
}

func (e *AdminAPIMutationScheduleAccountAnonymizationExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIMutationScheduleAccountAnonymizationExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPIMutationScheduleAccountAnonymizationExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIMutationScheduleAccountAnonymizationExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPIMutationScheduleAccountAnonymizationExecutedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *AdminAPIMutationScheduleAccountAnonymizationExecutedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &AdminAPIMutationScheduleAccountAnonymizationExecutedEventPayload{}

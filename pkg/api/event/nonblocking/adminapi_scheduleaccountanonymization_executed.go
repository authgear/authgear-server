package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AdminAPIScheduleAccountAnonymizationExecuted event.Type = "admin_api.schedule_account_anonymization.executed"
)

type AdminAPIScheduleAccountAnonymizationExecutedEventPayload struct {
	UserRef   model.UserRef `json:"-" resolve:"user"`
	UserModel model.User    `json:"user"`
}

func (e *AdminAPIScheduleAccountAnonymizationExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIScheduleAccountAnonymizationExecuted
}

func (e *AdminAPIScheduleAccountAnonymizationExecutedEventPayload) UserID() string {
	return e.UserModel.ID
}

func (e *AdminAPIScheduleAccountAnonymizationExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIScheduleAccountAnonymizationExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPIScheduleAccountAnonymizationExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIScheduleAccountAnonymizationExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPIScheduleAccountAnonymizationExecutedEventPayload) ReindexUserNeeded() bool {
	return false
}

func (e *AdminAPIScheduleAccountAnonymizationExecutedEventPayload) IsUserDeleted() bool {
	return false
}

var _ event.NonBlockingPayload = &AdminAPIScheduleAccountAnonymizationExecutedEventPayload{}

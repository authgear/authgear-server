package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AdminAPIMutationUnscheduleAccountAnonymizationExecuted event.Type = "admin_api.mutation.unschedule_account_anonymization.executed"
)

type AdminAPIMutationUnscheduleAccountAnonymizationExecutedEventPayload struct {
	UserRef   model.UserRef `json:"-" resolve:"user"`
	UserModel model.User    `json:"user"`
}

func (e *AdminAPIMutationUnscheduleAccountAnonymizationExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIMutationUnscheduleAccountAnonymizationExecuted
}

func (e *AdminAPIMutationUnscheduleAccountAnonymizationExecutedEventPayload) UserID() string {
	return e.UserModel.ID
}

func (e *AdminAPIMutationUnscheduleAccountAnonymizationExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIMutationUnscheduleAccountAnonymizationExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPIMutationUnscheduleAccountAnonymizationExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIMutationUnscheduleAccountAnonymizationExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPIMutationUnscheduleAccountAnonymizationExecutedEventPayload) RequireReindexUserIDs() []string {
	return []string{}
}

func (e *AdminAPIMutationUnscheduleAccountAnonymizationExecutedEventPayload) DeletedUserIDs() []string {
	return []string{}
}

var _ event.NonBlockingPayload = &AdminAPIMutationUnscheduleAccountAnonymizationExecutedEventPayload{}

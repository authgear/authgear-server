package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AdminAPIUnscheduleAccountAnonymizationExecuted event.Type = "admin_api.unschedule_account_anonymization.executed"
)

type AdminAPIUnscheduleAccountAnonymizationExecutedEventPayload struct {
	UserRef   model.UserRef `json:"-" resolve:"user"`
	UserModel model.User    `json:"user"`
}

func (e *AdminAPIUnscheduleAccountAnonymizationExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIUnscheduleAccountAnonymizationExecuted
}

func (e *AdminAPIUnscheduleAccountAnonymizationExecutedEventPayload) UserID() string {
	return e.UserModel.ID
}

func (e *AdminAPIUnscheduleAccountAnonymizationExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIUnscheduleAccountAnonymizationExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPIUnscheduleAccountAnonymizationExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIUnscheduleAccountAnonymizationExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPIUnscheduleAccountAnonymizationExecutedEventPayload) ReindexUserNeeded() bool {
	return false
}

func (e *AdminAPIUnscheduleAccountAnonymizationExecutedEventPayload) IsUserDeleted() bool {
	return false
}

var _ event.NonBlockingPayload = &AdminAPIUnscheduleAccountAnonymizationExecutedEventPayload{}

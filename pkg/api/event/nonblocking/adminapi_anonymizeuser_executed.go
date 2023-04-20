package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AdminAPIAnonymizeUserExecuted event.Type = "admin_api.anonymize_user.executed"
)

type AdminAPIAnonymizeUserExecutedEventPayload struct {
	UserRef   model.UserRef `json:"-" resolve:"user"`
	UserModel model.User    `json:"user"`
}

func (e *AdminAPIAnonymizeUserExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIAnonymizeUserExecuted
}

func (e *AdminAPIAnonymizeUserExecutedEventPayload) UserID() string {
	return e.UserModel.ID
}

func (e *AdminAPIAnonymizeUserExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIAnonymizeUserExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPIAnonymizeUserExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIAnonymizeUserExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPIAnonymizeUserExecutedEventPayload) ReindexUserNeeded() bool {
	return false
}

func (e *AdminAPIAnonymizeUserExecutedEventPayload) IsUserDeleted() bool {
	return false
}

var _ event.NonBlockingPayload = &AdminAPIAnonymizeUserExecutedEventPayload{}

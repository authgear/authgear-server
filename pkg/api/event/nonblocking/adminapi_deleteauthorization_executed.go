package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AdminAPIDeleteAuthorizationExecuted event.Type = "admin_api.delete_authorization.executed"
)

type AdminAPIDeleteAuthorizationExecutedEventPayload struct {
	Authorization model.Authorization `json:"authorization"`
}

func (e *AdminAPIDeleteAuthorizationExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIDeleteAuthorizationExecuted
}

func (e *AdminAPIDeleteAuthorizationExecutedEventPayload) UserID() string {
	return ""
}

func (e *AdminAPIDeleteAuthorizationExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIDeleteAuthorizationExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPIDeleteAuthorizationExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIDeleteAuthorizationExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPIDeleteAuthorizationExecutedEventPayload) ReindexUserNeeded() bool {
	return false
}

func (e *AdminAPIDeleteAuthorizationExecutedEventPayload) IsUserDeleted() bool {
	return false
}

var _ event.NonBlockingPayload = &AdminAPIDeleteAuthorizationExecutedEventPayload{}

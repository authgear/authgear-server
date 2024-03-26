package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AdminAPIMutationDeleteAuthorizationExecuted event.Type = "admin_api.mutation.delete_authorization.executed"
)

type AdminAPIMutationDeleteAuthorizationExecutedEventPayload struct {
	Authorization model.Authorization `json:"authorization"`
}

func (e *AdminAPIMutationDeleteAuthorizationExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIMutationDeleteAuthorizationExecuted
}

func (e *AdminAPIMutationDeleteAuthorizationExecutedEventPayload) UserID() string {
	return ""
}

func (e *AdminAPIMutationDeleteAuthorizationExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIMutationDeleteAuthorizationExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPIMutationDeleteAuthorizationExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIMutationDeleteAuthorizationExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPIMutationDeleteAuthorizationExecutedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *AdminAPIMutationDeleteAuthorizationExecutedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &AdminAPIMutationDeleteAuthorizationExecutedEventPayload{}

package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AdminAPIMutationDeleteResourceExecuted event.Type = "admin_api.mutation.delete_resource.executed"
)

type AdminAPIMutationDeleteResourceExecutedEventPayload struct {
	Resource model.Resource `json:"resource"`
}

func (e *AdminAPIMutationDeleteResourceExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIMutationDeleteResourceExecuted
}

func (e *AdminAPIMutationDeleteResourceExecutedEventPayload) UserID() string {
	return ""
}

func (e *AdminAPIMutationDeleteResourceExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIMutationDeleteResourceExecutedEventPayload) FillContext(ctx *event.Context) {}

func (e *AdminAPIMutationDeleteResourceExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIMutationDeleteResourceExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPIMutationDeleteResourceExecutedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *AdminAPIMutationDeleteResourceExecutedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &AdminAPIMutationDeleteResourceExecutedEventPayload{}

package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AdminAPIMutationCreateResourceExecuted event.Type = "admin_api.mutation.create_resource.executed"
)

type AdminAPIMutationCreateResourceExecutedEventPayload struct {
	Resource model.Resource `json:"resource"`
}

func (e *AdminAPIMutationCreateResourceExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIMutationCreateResourceExecuted
}

func (e *AdminAPIMutationCreateResourceExecutedEventPayload) UserID() string {
	return ""
}

func (e *AdminAPIMutationCreateResourceExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIMutationCreateResourceExecutedEventPayload) FillContext(ctx *event.Context) {}

func (e *AdminAPIMutationCreateResourceExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIMutationCreateResourceExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPIMutationCreateResourceExecutedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *AdminAPIMutationCreateResourceExecutedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &AdminAPIMutationCreateResourceExecutedEventPayload{}

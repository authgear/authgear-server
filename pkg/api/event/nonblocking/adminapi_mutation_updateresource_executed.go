package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AdminAPIMutationUpdateResourceExecuted event.Type = "admin_api.mutation.update_resource.executed"
)

type AdminAPIMutationUpdateResourceExecutedEventPayload struct {
	OriginalResource model.Resource `json:"original_resource"`
	NewResource      model.Resource `json:"new_resource"`
}

func (e *AdminAPIMutationUpdateResourceExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIMutationUpdateResourceExecuted
}

func (e *AdminAPIMutationUpdateResourceExecutedEventPayload) UserID() string {
	return ""
}

func (e *AdminAPIMutationUpdateResourceExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIMutationUpdateResourceExecutedEventPayload) FillContext(ctx *event.Context) {}

func (e *AdminAPIMutationUpdateResourceExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIMutationUpdateResourceExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPIMutationUpdateResourceExecutedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *AdminAPIMutationUpdateResourceExecutedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &AdminAPIMutationUpdateResourceExecutedEventPayload{}

package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AdminAPIMutationRemoveResourceFromClientIDExecuted event.Type = "admin_api.mutation.remove_resource_from_clientid.executed"
)

type AdminAPIMutationRemoveResourceFromClientIDExecutedEventPayload struct {
	Resource model.Resource `json:"resource"`
	ClientID string         `json:"client_id"`
}

func (e *AdminAPIMutationRemoveResourceFromClientIDExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIMutationRemoveResourceFromClientIDExecuted
}

func (e *AdminAPIMutationRemoveResourceFromClientIDExecutedEventPayload) UserID() string {
	return ""
}

func (e *AdminAPIMutationRemoveResourceFromClientIDExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIMutationRemoveResourceFromClientIDExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPIMutationRemoveResourceFromClientIDExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIMutationRemoveResourceFromClientIDExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPIMutationRemoveResourceFromClientIDExecutedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *AdminAPIMutationRemoveResourceFromClientIDExecutedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &AdminAPIMutationRemoveResourceFromClientIDExecutedEventPayload{}

package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AdminAPIMutationAddResourceToClientIDExecuted event.Type = "admin_api.mutation.add_resource_to_clientid.executed"
)

type AdminAPIMutationAddResourceToClientIDExecutedEventPayload struct {
	Resource model.Resource `json:"resource"`
	ClientID string         `json:"client_id"`
}

func (e *AdminAPIMutationAddResourceToClientIDExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIMutationAddResourceToClientIDExecuted
}

func (e *AdminAPIMutationAddResourceToClientIDExecutedEventPayload) UserID() string {
	return ""
}

func (e *AdminAPIMutationAddResourceToClientIDExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIMutationAddResourceToClientIDExecutedEventPayload) FillContext(ctx *event.Context) {}

func (e *AdminAPIMutationAddResourceToClientIDExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIMutationAddResourceToClientIDExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPIMutationAddResourceToClientIDExecutedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *AdminAPIMutationAddResourceToClientIDExecutedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &AdminAPIMutationAddResourceToClientIDExecutedEventPayload{}

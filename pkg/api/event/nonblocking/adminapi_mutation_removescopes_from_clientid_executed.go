package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AdminAPIMutationRemoveScopesFromClientIDExecuted event.Type = "admin_api.mutation.remove_scopes_from_clientid.executed"
)

type AdminAPIMutationRemoveScopesFromClientIDExecutedEventPayload struct {
	ResourceURI string         `json:"resource_uri"`
	ClientID    string         `json:"client_id"`
	Scopes      []*model.Scope `json:"scopes"`
}

func (e *AdminAPIMutationRemoveScopesFromClientIDExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIMutationRemoveScopesFromClientIDExecuted
}

func (e *AdminAPIMutationRemoveScopesFromClientIDExecutedEventPayload) UserID() string {
	return ""
}

func (e *AdminAPIMutationRemoveScopesFromClientIDExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIMutationRemoveScopesFromClientIDExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPIMutationRemoveScopesFromClientIDExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIMutationRemoveScopesFromClientIDExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPIMutationRemoveScopesFromClientIDExecutedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *AdminAPIMutationRemoveScopesFromClientIDExecutedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &AdminAPIMutationRemoveScopesFromClientIDExecutedEventPayload{}

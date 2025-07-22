package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AdminAPIMutationAddScopesToClientIDExecuted event.Type = "admin_api.mutation.add_scopes_to_clientid.executed"
)

type AdminAPIMutationAddScopesToClientIDExecutedEventPayload struct {
	ResourceURI string         `json:"resource_uri"`
	ClientID    string         `json:"client_id"`
	Scopes      []*model.Scope `json:"scopes"`
}

func (e *AdminAPIMutationAddScopesToClientIDExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIMutationAddScopesToClientIDExecuted
}

func (e *AdminAPIMutationAddScopesToClientIDExecutedEventPayload) UserID() string {
	return ""
}

func (e *AdminAPIMutationAddScopesToClientIDExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIMutationAddScopesToClientIDExecutedEventPayload) FillContext(ctx *event.Context) {}

func (e *AdminAPIMutationAddScopesToClientIDExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIMutationAddScopesToClientIDExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPIMutationAddScopesToClientIDExecutedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *AdminAPIMutationAddScopesToClientIDExecutedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &AdminAPIMutationAddScopesToClientIDExecutedEventPayload{}

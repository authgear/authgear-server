package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AdminAPIMutationReplaceScopesOfClientIDExecuted event.Type = "admin_api.mutation.replace_scopes_of_clientid.executed"
)

type AdminAPIMutationReplaceScopesOfClientIDExecutedEventPayload struct {
	ResourceURI string         `json:"resource_uri"`
	ClientID    string         `json:"client_id"`
	Scopes      []*model.Scope `json:"scopes"`
}

func (e *AdminAPIMutationReplaceScopesOfClientIDExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIMutationReplaceScopesOfClientIDExecuted
}

func (e *AdminAPIMutationReplaceScopesOfClientIDExecutedEventPayload) UserID() string {
	return ""
}

func (e *AdminAPIMutationReplaceScopesOfClientIDExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIMutationReplaceScopesOfClientIDExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPIMutationReplaceScopesOfClientIDExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIMutationReplaceScopesOfClientIDExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPIMutationReplaceScopesOfClientIDExecutedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *AdminAPIMutationReplaceScopesOfClientIDExecutedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &AdminAPIMutationReplaceScopesOfClientIDExecutedEventPayload{}

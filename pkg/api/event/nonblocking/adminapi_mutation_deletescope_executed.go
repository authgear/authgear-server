package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AdminAPIMutationDeleteScopeExecuted event.Type = "admin_api.mutation.delete_scope.executed"
)

type AdminAPIMutationDeleteScopeExecutedEventPayload struct {
	Scope model.Scope `json:"scope"`
}

func (e *AdminAPIMutationDeleteScopeExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIMutationDeleteScopeExecuted
}

func (e *AdminAPIMutationDeleteScopeExecutedEventPayload) UserID() string {
	return ""
}

func (e *AdminAPIMutationDeleteScopeExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIMutationDeleteScopeExecutedEventPayload) FillContext(ctx *event.Context) {}

func (e *AdminAPIMutationDeleteScopeExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIMutationDeleteScopeExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPIMutationDeleteScopeExecutedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *AdminAPIMutationDeleteScopeExecutedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &AdminAPIMutationDeleteScopeExecutedEventPayload{}

package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AdminAPIMutationCreateScopeExecuted event.Type = "admin_api.mutation.create_scope.executed"
)

type AdminAPIMutationCreateScopeExecutedEventPayload struct {
	Scope model.Scope `json:"scope"`
}

func (e *AdminAPIMutationCreateScopeExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIMutationCreateScopeExecuted
}

func (e *AdminAPIMutationCreateScopeExecutedEventPayload) UserID() string {
	return ""
}

func (e *AdminAPIMutationCreateScopeExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIMutationCreateScopeExecutedEventPayload) FillContext(ctx *event.Context) {}

func (e *AdminAPIMutationCreateScopeExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIMutationCreateScopeExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPIMutationCreateScopeExecutedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *AdminAPIMutationCreateScopeExecutedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &AdminAPIMutationCreateScopeExecutedEventPayload{}

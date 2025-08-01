package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AdminAPIMutationUpdateScopeExecuted event.Type = "admin_api.mutation.update_scope.executed"
)

type AdminAPIMutationUpdateScopeExecutedEventPayload struct {
	OriginalScope model.Scope `json:"original_scope"`
	NewScope      model.Scope `json:"new_scope"`
}

func (e *AdminAPIMutationUpdateScopeExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIMutationUpdateScopeExecuted
}

func (e *AdminAPIMutationUpdateScopeExecutedEventPayload) UserID() string {
	return ""
}

func (e *AdminAPIMutationUpdateScopeExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIMutationUpdateScopeExecutedEventPayload) FillContext(ctx *event.Context) {}

func (e *AdminAPIMutationUpdateScopeExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIMutationUpdateScopeExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPIMutationUpdateScopeExecutedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *AdminAPIMutationUpdateScopeExecutedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &AdminAPIMutationUpdateScopeExecutedEventPayload{}

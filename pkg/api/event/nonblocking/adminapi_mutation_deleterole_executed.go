package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
)

const (
	AdminAPIMutationDeleteRoleExecuted event.Type = "admin_api.mutation.delete_role.executed"
)

type AdminAPIMutationDeleteRoleExecutedEventPayload struct {
	AffectedUserIDs []string `json:"-"`
}

func (e *AdminAPIMutationDeleteRoleExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIMutationDeleteRoleExecuted
}

func (e *AdminAPIMutationDeleteRoleExecutedEventPayload) UserID() string {
	return ""
}

func (e *AdminAPIMutationDeleteRoleExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIMutationDeleteRoleExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPIMutationDeleteRoleExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIMutationDeleteRoleExecutedEventPayload) ForAudit() bool {
	// FIXME(tung): Should be true
	return false
}

func (e *AdminAPIMutationDeleteRoleExecutedEventPayload) RequireReindexUserIDs() []string {
	return e.AffectedUserIDs
}

func (e *AdminAPIMutationDeleteRoleExecutedEventPayload) DeletedUserIDs() []string {
	return []string{}
}

var _ event.NonBlockingPayload = &AdminAPIMutationDeleteRoleExecutedEventPayload{}

package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
)

const (
	AdminAPIMutationRemoveRoleFromGroupsExecuted event.Type = "admin_api.mutation.remove_role_from_groups.executed"
)

type AdminAPIMutationRemoveRoleFromGroupsExecutedEventPayload struct {
	AffectedUserIDs []string `json:"-"`
	RoleID          string   `json:"role_id"`
	GroupIDs        []string `json:"group_ids"`
}

func (e *AdminAPIMutationRemoveRoleFromGroupsExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIMutationRemoveRoleFromGroupsExecuted
}

func (e *AdminAPIMutationRemoveRoleFromGroupsExecutedEventPayload) UserID() string {
	return ""
}

func (e *AdminAPIMutationRemoveRoleFromGroupsExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIMutationRemoveRoleFromGroupsExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPIMutationRemoveRoleFromGroupsExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIMutationRemoveRoleFromGroupsExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPIMutationRemoveRoleFromGroupsExecutedEventPayload) RequireReindexUserIDs() []string {
	return e.AffectedUserIDs
}

func (e *AdminAPIMutationRemoveRoleFromGroupsExecutedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &AdminAPIMutationRemoveRoleFromGroupsExecutedEventPayload{}

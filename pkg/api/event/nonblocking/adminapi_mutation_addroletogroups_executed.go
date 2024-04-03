package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
)

const (
	AdminAPIMutationAddRoleToGroupsExecuted event.Type = "admin_api.mutation.add_role_to_groups.executed"
)

type AdminAPIMutationAddRoleToGroupsExecutedEventPayload struct {
	AffectedUserIDs []string `json:"-"`
	RoleID          string   `json:"role_id"`
	GroupIDs        []string `json:"group_ids"`
}

func (e *AdminAPIMutationAddRoleToGroupsExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIMutationAddRoleToGroupsExecuted
}

func (e *AdminAPIMutationAddRoleToGroupsExecutedEventPayload) UserID() string {
	return ""
}

func (e *AdminAPIMutationAddRoleToGroupsExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIMutationAddRoleToGroupsExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPIMutationAddRoleToGroupsExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIMutationAddRoleToGroupsExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPIMutationAddRoleToGroupsExecutedEventPayload) RequireReindexUserIDs() []string {
	return e.AffectedUserIDs
}

func (e *AdminAPIMutationAddRoleToGroupsExecutedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &AdminAPIMutationAddRoleToGroupsExecutedEventPayload{}

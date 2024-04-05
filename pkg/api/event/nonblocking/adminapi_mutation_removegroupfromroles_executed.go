package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
)

const (
	AdminAPIMutationRemoveGroupFromRolesExecuted event.Type = "admin_api.mutation.remove_group_from_roles.executed"
)

type AdminAPIMutationRemoveGroupFromRolesExecutedEventPayload struct {
	AffectedUserIDs []string `json:"-"`
	GroupID         string   `json:"group_id"`
	RoleIDs         []string `json:"role_ids"`
}

func (e *AdminAPIMutationRemoveGroupFromRolesExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIMutationRemoveGroupFromRolesExecuted
}

func (e *AdminAPIMutationRemoveGroupFromRolesExecutedEventPayload) UserID() string {
	return ""
}

func (e *AdminAPIMutationRemoveGroupFromRolesExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIMutationRemoveGroupFromRolesExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPIMutationRemoveGroupFromRolesExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIMutationRemoveGroupFromRolesExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPIMutationRemoveGroupFromRolesExecutedEventPayload) RequireReindexUserIDs() []string {
	return e.AffectedUserIDs
}

func (e *AdminAPIMutationRemoveGroupFromRolesExecutedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &AdminAPIMutationRemoveGroupFromRolesExecutedEventPayload{}

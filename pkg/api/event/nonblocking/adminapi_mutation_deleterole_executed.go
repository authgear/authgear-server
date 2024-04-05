package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AdminAPIMutationDeleteRoleExecuted event.Type = "admin_api.mutation.delete_role.executed"
)

type AdminAPIMutationDeleteRoleExecutedEventPayload struct {
	AffectedUserIDs []string   `json:"-"`
	Role            model.Role `json:"role"`
	RoleGroupIDs    []string   `json:"role_group_ids"`
	RoleUserIDs     []string   `json:"role_user_ids"`
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
	return true
}

func (e *AdminAPIMutationDeleteRoleExecutedEventPayload) RequireReindexUserIDs() []string {
	return e.AffectedUserIDs
}

func (e *AdminAPIMutationDeleteRoleExecutedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &AdminAPIMutationDeleteRoleExecutedEventPayload{}

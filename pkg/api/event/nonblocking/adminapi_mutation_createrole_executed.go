package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AdminAPIMutationCreateRoleExecuted event.Type = "admin_api.mutation.create_role.executed"
)

type AdminAPIMutationCreateRoleExecutedEventPayload struct {
	Role model.Role `json:"role"`
}

func (e *AdminAPIMutationCreateRoleExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIMutationCreateRoleExecuted
}

func (e *AdminAPIMutationCreateRoleExecutedEventPayload) UserID() string {
	return ""
}

func (e *AdminAPIMutationCreateRoleExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIMutationCreateRoleExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPIMutationCreateRoleExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIMutationCreateRoleExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPIMutationCreateRoleExecutedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *AdminAPIMutationCreateRoleExecutedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &AdminAPIMutationCreateRoleExecutedEventPayload{}

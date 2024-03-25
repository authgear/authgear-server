package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
)

const (
	AdminAPIMutationAddRoleToUsersExecuted event.Type = "admin_api.mutation.add_role_to_users.executed"
)

type AdminAPIMutationAddRoleToUsersExecutedEventPayload struct {
	AffectedUserIDs []string `json:"-"`
}

func (e *AdminAPIMutationAddRoleToUsersExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIMutationAddRoleToUsersExecuted
}

func (e *AdminAPIMutationAddRoleToUsersExecutedEventPayload) UserID() string {
	return ""
}

func (e *AdminAPIMutationAddRoleToUsersExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIMutationAddRoleToUsersExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPIMutationAddRoleToUsersExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIMutationAddRoleToUsersExecutedEventPayload) ForAudit() bool {
	// FIXME(tung): Should be true
	return false
}

func (e *AdminAPIMutationAddRoleToUsersExecutedEventPayload) RequireReindexUserIDs() []string {
	return e.AffectedUserIDs
}

func (e *AdminAPIMutationAddRoleToUsersExecutedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &AdminAPIMutationAddRoleToUsersExecutedEventPayload{}

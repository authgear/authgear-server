package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
)

const (
	AdminAPIMutationRemoveUserFromRolesExecuted event.Type = "admin_api.mutation.remove_user_from_roles.executed"
)

type AdminAPIMutationRemoveUserFromRolesExecutedEventPayload struct {
	AffectedUserIDs []string `json:"-"`
}

func (e *AdminAPIMutationRemoveUserFromRolesExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIMutationRemoveUserFromRolesExecuted
}

func (e *AdminAPIMutationRemoveUserFromRolesExecutedEventPayload) UserID() string {
	return ""
}

func (e *AdminAPIMutationRemoveUserFromRolesExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIMutationRemoveUserFromRolesExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPIMutationRemoveUserFromRolesExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIMutationRemoveUserFromRolesExecutedEventPayload) ForAudit() bool {
	// FIXME(tung): Should be true
	return false
}

func (e *AdminAPIMutationRemoveUserFromRolesExecutedEventPayload) RequireReindexUserIDs() []string {
	return e.AffectedUserIDs
}

func (e *AdminAPIMutationRemoveUserFromRolesExecutedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &AdminAPIMutationRemoveUserFromRolesExecutedEventPayload{}

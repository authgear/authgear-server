package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
)

const (
	AdminAPIMutationAddUserToRolesExecuted event.Type = "admin_api.mutation.add_user_to_roles.executed"
)

type AdminAPIMutationAddUserToRolesExecutedEventPayload struct {
	AffectedUserIDs []string `json:"-"`
}

func (e *AdminAPIMutationAddUserToRolesExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIMutationAddUserToRolesExecuted
}

func (e *AdminAPIMutationAddUserToRolesExecutedEventPayload) UserID() string {
	return ""
}

func (e *AdminAPIMutationAddUserToRolesExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIMutationAddUserToRolesExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPIMutationAddUserToRolesExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIMutationAddUserToRolesExecutedEventPayload) ForAudit() bool {
	// FIXME(tung): Should be true
	return false
}

func (e *AdminAPIMutationAddUserToRolesExecutedEventPayload) RequireReindexUserIDs() []string {
	return e.AffectedUserIDs
}

func (e *AdminAPIMutationAddUserToRolesExecutedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &AdminAPIMutationAddUserToRolesExecutedEventPayload{}

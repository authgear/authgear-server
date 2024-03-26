package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
)

const (
	AdminAPIMutationUpdateRoleExecuted event.Type = "admin_api.mutation.update_role.executed"
)

type AdminAPIMutationUpdateRoleExecutedEventPayload struct {
	AffectedUserIDs []string `json:"-"`
}

func (e *AdminAPIMutationUpdateRoleExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIMutationUpdateRoleExecuted
}

func (e *AdminAPIMutationUpdateRoleExecutedEventPayload) UserID() string {
	return ""
}

func (e *AdminAPIMutationUpdateRoleExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIMutationUpdateRoleExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPIMutationUpdateRoleExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIMutationUpdateRoleExecutedEventPayload) ForAudit() bool {
	// FIXME(tung): Should be true
	return false
}

func (e *AdminAPIMutationUpdateRoleExecutedEventPayload) RequireReindexUserIDs() []string {
	return e.AffectedUserIDs
}

func (e *AdminAPIMutationUpdateRoleExecutedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &AdminAPIMutationUpdateRoleExecutedEventPayload{}

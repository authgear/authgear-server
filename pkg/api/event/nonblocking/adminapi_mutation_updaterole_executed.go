package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AdminAPIMutationUpdateRoleExecuted event.Type = "admin_api.mutation.update_role.executed"
)

type AdminAPIMutationUpdateRoleExecutedEventPayload struct {
	AffectedUserIDs []string   `json:"-"`
	OriginalRole    model.Role `json:"original_role"`
	NewRole         model.Role `json:"new_role"`
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
	return true
}

func (e *AdminAPIMutationUpdateRoleExecutedEventPayload) RequireReindexUserIDs() []string {
	return e.AffectedUserIDs
}

func (e *AdminAPIMutationUpdateRoleExecutedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &AdminAPIMutationUpdateRoleExecutedEventPayload{}

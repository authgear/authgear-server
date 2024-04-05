package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AdminAPIMutationDeleteGroupExecuted event.Type = "admin_api.mutation.delete_group.executed"
)

type AdminAPIMutationDeleteGroupExecutedEventPayload struct {
	Group        model.Group `json:"group"`
	GroupRoleIDs []string    `json:"group_role_ids"`
	GroupUserIDs []string    `json:"group_user_ids"`
}

func (e *AdminAPIMutationDeleteGroupExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIMutationDeleteGroupExecuted
}

func (e *AdminAPIMutationDeleteGroupExecutedEventPayload) UserID() string {
	return ""
}

func (e *AdminAPIMutationDeleteGroupExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIMutationDeleteGroupExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPIMutationDeleteGroupExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIMutationDeleteGroupExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPIMutationDeleteGroupExecutedEventPayload) RequireReindexUserIDs() []string {
	return e.GroupUserIDs
}

func (e *AdminAPIMutationDeleteGroupExecutedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &AdminAPIMutationDeleteGroupExecutedEventPayload{}

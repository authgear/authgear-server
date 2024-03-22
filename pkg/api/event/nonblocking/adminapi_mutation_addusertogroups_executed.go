package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
)

const (
	AdminAPIMutationAddUserToGroupsExecuted event.Type = "admin_api.mutation.add_user_to_groups.executed"
)

type AdminAPIMutationAddUserToGroupsExecutedEventPayload struct {
	AffectedUserIDs []string `json:"-"`
}

func (e *AdminAPIMutationAddUserToGroupsExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIMutationAddUserToGroupsExecuted
}

func (e *AdminAPIMutationAddUserToGroupsExecutedEventPayload) UserID() string {
	return ""
}

func (e *AdminAPIMutationAddUserToGroupsExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIMutationAddUserToGroupsExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPIMutationAddUserToGroupsExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIMutationAddUserToGroupsExecutedEventPayload) ForAudit() bool {
	// FIXME(tung): Should be true
	return false
}

func (e *AdminAPIMutationAddUserToGroupsExecutedEventPayload) RequireReindexUserIDs() []string {
	return e.AffectedUserIDs
}

func (e *AdminAPIMutationAddUserToGroupsExecutedEventPayload) DeletedUserIDs() []string {
	return []string{}
}

var _ event.NonBlockingPayload = &AdminAPIMutationAddUserToGroupsExecutedEventPayload{}

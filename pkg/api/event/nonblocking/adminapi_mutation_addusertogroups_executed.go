package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
)

const (
	AdminAPIMutationAddUserToGroupsExecuted event.Type = "admin_api.mutation.add_user_to_groups.executed"
)

type AdminAPIMutationAddUserToGroupsExecutedEventPayload struct {
	UserID_  string   `json:"user_id"`
	GroupIDs []string `json:"group_ids"`
}

func (e *AdminAPIMutationAddUserToGroupsExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIMutationAddUserToGroupsExecuted
}

func (e *AdminAPIMutationAddUserToGroupsExecutedEventPayload) UserID() string {
	return e.UserID_
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
	return true
}

func (e *AdminAPIMutationAddUserToGroupsExecutedEventPayload) RequireReindexUserIDs() []string {
	return []string{e.UserID_}
}

func (e *AdminAPIMutationAddUserToGroupsExecutedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &AdminAPIMutationAddUserToGroupsExecutedEventPayload{}

package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
)

const (
	AdminAPIMutationAddGroupToUsersExecuted event.Type = "admin_api.mutation.add_group_to_users.executed"
)

type AdminAPIMutationAddGroupToUsersExecutedEventPayload struct {
	UserIDs []string `json:"user_ids"`
	GroupID string   `json:"group_id"`
}

func (e *AdminAPIMutationAddGroupToUsersExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIMutationAddGroupToUsersExecuted
}

func (e *AdminAPIMutationAddGroupToUsersExecutedEventPayload) UserID() string {
	return ""
}

func (e *AdminAPIMutationAddGroupToUsersExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIMutationAddGroupToUsersExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPIMutationAddGroupToUsersExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIMutationAddGroupToUsersExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPIMutationAddGroupToUsersExecutedEventPayload) RequireReindexUserIDs() []string {
	return e.UserIDs
}

func (e *AdminAPIMutationAddGroupToUsersExecutedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &AdminAPIMutationAddGroupToUsersExecutedEventPayload{}

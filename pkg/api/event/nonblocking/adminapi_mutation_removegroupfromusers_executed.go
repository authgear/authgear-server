package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
)

const (
	AdminAPIMutationRemoveGroupFromUsersExecuted event.Type = "admin_api.mutation.remove_group_from_users.executed"
)

type AdminAPIMutationRemoveGroupFromUsersExecutedEventPayload struct {
	UserIDs []string `json:"user_ids"`
	GroupID string   `json:"group_id"`
}

func (e *AdminAPIMutationRemoveGroupFromUsersExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIMutationRemoveGroupFromUsersExecuted
}

func (e *AdminAPIMutationRemoveGroupFromUsersExecutedEventPayload) UserID() string {
	return ""
}

func (e *AdminAPIMutationRemoveGroupFromUsersExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIMutationRemoveGroupFromUsersExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPIMutationRemoveGroupFromUsersExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIMutationRemoveGroupFromUsersExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPIMutationRemoveGroupFromUsersExecutedEventPayload) RequireReindexUserIDs() []string {
	return e.UserIDs
}

func (e *AdminAPIMutationRemoveGroupFromUsersExecutedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &AdminAPIMutationRemoveGroupFromUsersExecutedEventPayload{}

package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
)

const (
	AdminAPIMutationRemoveUserFromGroupsExecuted event.Type = "admin_api.mutation.remove_user_from_groups.executed"
)

type AdminAPIMutationRemoveUserFromGroupsExecutedEventPayload struct {
	UserID_  string   `json:"user_id"`
	GroupIDs []string `json:"group_ids"`
}

func (e *AdminAPIMutationRemoveUserFromGroupsExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIMutationRemoveUserFromGroupsExecuted
}

func (e *AdminAPIMutationRemoveUserFromGroupsExecutedEventPayload) UserID() string {
	return e.UserID_
}

func (e *AdminAPIMutationRemoveUserFromGroupsExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIMutationRemoveUserFromGroupsExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPIMutationRemoveUserFromGroupsExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIMutationRemoveUserFromGroupsExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPIMutationRemoveUserFromGroupsExecutedEventPayload) RequireReindexUserIDs() []string {
	return []string{e.UserID_}
}

func (e *AdminAPIMutationRemoveUserFromGroupsExecutedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &AdminAPIMutationRemoveUserFromGroupsExecutedEventPayload{}

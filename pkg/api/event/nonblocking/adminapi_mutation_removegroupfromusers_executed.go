package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
)

const (
	AdminAPIMutationRemoveGroupFromUsersExecuted event.Type = "admin_api.mutation.remove_group_from_users.executed"
)

type AdminAPIMutationRemoveGroupFromUsersExecutedEventPayload struct {
	AffectedUserIDs []string `json:"-"`
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
	// FIXME(tung): Should be true
	return false
}

func (e *AdminAPIMutationRemoveGroupFromUsersExecutedEventPayload) RequireReindexUserIDs() []string {
	return e.AffectedUserIDs
}

func (e *AdminAPIMutationRemoveGroupFromUsersExecutedEventPayload) DeletedUserIDs() []string {
	return []string{}
}

var _ event.NonBlockingPayload = &AdminAPIMutationRemoveGroupFromUsersExecutedEventPayload{}

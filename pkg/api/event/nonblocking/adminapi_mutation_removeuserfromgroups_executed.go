package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
)

const (
	AdminAPIMutationRemoveUserFromGroupsExecuted event.Type = "admin_api.mutation.remove_user_from_groups.executed"
)

type AdminAPIMutationRemoveUserFromGroupsExecutedEventPayload struct {
	AffectedUserIDs []string `json:"-"`
}

func (e *AdminAPIMutationRemoveUserFromGroupsExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIMutationRemoveUserFromGroupsExecuted
}

func (e *AdminAPIMutationRemoveUserFromGroupsExecutedEventPayload) UserID() string {
	return ""
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
	// FIXME(tung): Should be true
	return false
}

func (e *AdminAPIMutationRemoveUserFromGroupsExecutedEventPayload) RequireReindexUserIDs() []string {
	return e.AffectedUserIDs
}

func (e *AdminAPIMutationRemoveUserFromGroupsExecutedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &AdminAPIMutationRemoveUserFromGroupsExecutedEventPayload{}

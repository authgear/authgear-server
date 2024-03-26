package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
)

const (
	AdminAPIMutationAddGroupToUsersExecuted event.Type = "admin_api.mutation.add_group_to_users.executed"
)

type AdminAPIMutationAddGroupToUsersExecutedEventPayload struct {
	AffectedUserIDs []string `json:"-"`
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
	// FIXME(tung): Should be true
	return false
}

func (e *AdminAPIMutationAddGroupToUsersExecutedEventPayload) RequireReindexUserIDs() []string {
	return e.AffectedUserIDs
}

func (e *AdminAPIMutationAddGroupToUsersExecutedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &AdminAPIMutationAddGroupToUsersExecutedEventPayload{}

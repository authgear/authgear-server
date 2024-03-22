package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
)

const (
	AdminAPIMutationDeleteGroupExecuted event.Type = "admin_api.mutation.delete_group.executed"
)

type AdminAPIMutationDeleteGroupExecutedEventPayload struct {
	AffectedUserIDs []string `json:"-"`
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
	// FIXME(tung): Should be true
	return false
}

func (e *AdminAPIMutationDeleteGroupExecutedEventPayload) RequireReindexUserIDs() []string {
	return e.AffectedUserIDs
}

func (e *AdminAPIMutationDeleteGroupExecutedEventPayload) DeletedUserIDs() []string {
	return []string{}
}

var _ event.NonBlockingPayload = &AdminAPIMutationDeleteGroupExecutedEventPayload{}

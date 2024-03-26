package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
)

const (
	AdminAPIMutationUpdateGroupExecuted event.Type = "admin_api.mutation.update_group.executed"
)

type AdminAPIMutationUpdateGroupExecutedEventPayload struct {
	AffectedUserIDs []string `json:"-"`
}

func (e *AdminAPIMutationUpdateGroupExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIMutationUpdateGroupExecuted
}

func (e *AdminAPIMutationUpdateGroupExecutedEventPayload) UserID() string {
	return ""
}

func (e *AdminAPIMutationUpdateGroupExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIMutationUpdateGroupExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPIMutationUpdateGroupExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIMutationUpdateGroupExecutedEventPayload) ForAudit() bool {
	// FIXME(tung): Should be true
	return false
}

func (e *AdminAPIMutationUpdateGroupExecutedEventPayload) RequireReindexUserIDs() []string {
	return e.AffectedUserIDs
}

func (e *AdminAPIMutationUpdateGroupExecutedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &AdminAPIMutationUpdateGroupExecutedEventPayload{}

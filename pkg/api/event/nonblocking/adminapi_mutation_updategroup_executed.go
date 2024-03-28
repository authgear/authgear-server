package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AdminAPIMutationUpdateGroupExecuted event.Type = "admin_api.mutation.update_group.executed"
)

type AdminAPIMutationUpdateGroupExecutedEventPayload struct {
	AffectedUserIDs []string    `json:"-"`
	OriginalGroup   model.Group `json:"original_group"`
	NewGroup        model.Group `json:"new_group"`
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
	return true
}

func (e *AdminAPIMutationUpdateGroupExecutedEventPayload) RequireReindexUserIDs() []string {
	return e.AffectedUserIDs
}

func (e *AdminAPIMutationUpdateGroupExecutedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &AdminAPIMutationUpdateGroupExecutedEventPayload{}

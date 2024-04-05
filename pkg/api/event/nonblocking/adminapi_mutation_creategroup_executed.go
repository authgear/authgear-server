package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AdminAPIMutationCreateGroupExecuted event.Type = "admin_api.mutation.create_group.executed"
)

type AdminAPIMutationCreateGroupExecutedEventPayload struct {
	Group model.Group `json:"group"`
}

func (e *AdminAPIMutationCreateGroupExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIMutationCreateGroupExecuted
}

func (e *AdminAPIMutationCreateGroupExecutedEventPayload) UserID() string {
	return ""
}

func (e *AdminAPIMutationCreateGroupExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIMutationCreateGroupExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPIMutationCreateGroupExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIMutationCreateGroupExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPIMutationCreateGroupExecutedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *AdminAPIMutationCreateGroupExecutedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &AdminAPIMutationCreateGroupExecutedEventPayload{}

package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AdminAPIMutationCreateSessionExecuted event.Type = "admin_api.mutation.create_session.executed"
)

type AdminAPIMutationCreateSessionExecutedEventPayload struct {
	UserRef   model.UserRef `json:"-" resolve:"user"`
	UserModel model.User    `json:"user"`
	Session   model.Session `json:"session"`
}

func (e *AdminAPIMutationCreateSessionExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIMutationCreateSessionExecuted
}

func (e *AdminAPIMutationCreateSessionExecutedEventPayload) UserID() string {
	return e.UserModel.ID
}

func (e *AdminAPIMutationCreateSessionExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIMutationCreateSessionExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPIMutationCreateSessionExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIMutationCreateSessionExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPIMutationCreateSessionExecutedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *AdminAPIMutationCreateSessionExecutedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &AdminAPIMutationCreateSessionExecutedEventPayload{}

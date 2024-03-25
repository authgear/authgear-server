package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AdminAPIMutationCreateUserExecuted event.Type = "admin_api.mutation.create_user.executed"
)

type AdminAPIMutationCreateUserExecutedEventPayload struct {
	UserRef   model.UserRef `json:"-" resolve:"user"`
	UserModel model.User    `json:"user"`
}

func (e *AdminAPIMutationCreateUserExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIMutationCreateUserExecuted
}

func (e *AdminAPIMutationCreateUserExecutedEventPayload) UserID() string {
	return e.UserModel.ID
}

func (e *AdminAPIMutationCreateUserExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIMutationCreateUserExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPIMutationCreateUserExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIMutationCreateUserExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPIMutationCreateUserExecutedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *AdminAPIMutationCreateUserExecutedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &AdminAPIMutationCreateUserExecutedEventPayload{}

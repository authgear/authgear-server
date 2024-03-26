package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AdminAPIMutationDeleteUserExecuted event.Type = "admin_api.mutation.delete_user.executed"
)

type AdminAPIMutationDeleteUserExecutedEventPayload struct {
	UserModel model.User `json:"user"`
}

func (e *AdminAPIMutationDeleteUserExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIMutationDeleteUserExecuted
}

func (e *AdminAPIMutationDeleteUserExecutedEventPayload) UserID() string {
	return e.UserModel.ID
}

func (e *AdminAPIMutationDeleteUserExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIMutationDeleteUserExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPIMutationDeleteUserExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIMutationDeleteUserExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPIMutationDeleteUserExecutedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *AdminAPIMutationDeleteUserExecutedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &AdminAPIMutationDeleteUserExecutedEventPayload{}

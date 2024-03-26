package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AdminAPIMutationUpdateUserExecuted event.Type = "admin_api.mutation.update_user.executed"
)

type AdminAPIMutationUpdateUserExecutedEventPayload struct {
	UserRef   model.UserRef `json:"-" resolve:"user"`
	UserModel model.User    `json:"user"`
}

func (e *AdminAPIMutationUpdateUserExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIMutationUpdateUserExecuted
}

func (e *AdminAPIMutationUpdateUserExecutedEventPayload) UserID() string {
	return e.UserModel.ID
}

func (e *AdminAPIMutationUpdateUserExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIMutationUpdateUserExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPIMutationUpdateUserExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIMutationUpdateUserExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPIMutationUpdateUserExecutedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *AdminAPIMutationUpdateUserExecutedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &AdminAPIMutationUpdateUserExecutedEventPayload{}

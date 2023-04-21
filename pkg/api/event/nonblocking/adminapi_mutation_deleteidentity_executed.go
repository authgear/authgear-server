package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AdminAPIMutationDeleteIdentityExecuted event.Type = "admin_api.mutation.delete_identity.executed"
)

type AdminAPIMutationDeleteIdentityExecutedEventPayload struct {
	UserRef   model.UserRef  `json:"-" resolve:"user"`
	UserModel model.User     `json:"user"`
	Identity  model.Identity `json:"identity"`
}

func (e *AdminAPIMutationDeleteIdentityExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIMutationDeleteIdentityExecuted
}

func (e *AdminAPIMutationDeleteIdentityExecutedEventPayload) UserID() string {
	return e.UserModel.ID
}

func (e *AdminAPIMutationDeleteIdentityExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIMutationDeleteIdentityExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPIMutationDeleteIdentityExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIMutationDeleteIdentityExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPIMutationDeleteIdentityExecutedEventPayload) ReindexUserNeeded() bool {
	return false
}

func (e *AdminAPIMutationDeleteIdentityExecutedEventPayload) IsUserDeleted() bool {
	return false
}

var _ event.NonBlockingPayload = &AdminAPIMutationDeleteIdentityExecutedEventPayload{}

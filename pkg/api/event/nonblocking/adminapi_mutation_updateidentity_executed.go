package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AdminAPIMutationUpdateIdentityExecuted event.Type = "admin_api.mutation.update_identity.executed"
)

type AdminAPIMutationUpdateIdentityExecutedEventPayload struct {
	UserRef   model.UserRef  `json:"-" resolve:"user"`
	UserModel model.User     `json:"user"`
	Identity  model.Identity `json:"identity"`
}

func (e *AdminAPIMutationUpdateIdentityExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIMutationUpdateIdentityExecuted
}

func (e *AdminAPIMutationUpdateIdentityExecutedEventPayload) UserID() string {
	return e.UserModel.ID
}

func (e *AdminAPIMutationUpdateIdentityExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIMutationUpdateIdentityExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPIMutationUpdateIdentityExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIMutationUpdateIdentityExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPIMutationUpdateIdentityExecutedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *AdminAPIMutationUpdateIdentityExecutedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &AdminAPIMutationUpdateIdentityExecutedEventPayload{}

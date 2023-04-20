package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AdminAPIDeleteIdentityExecuted event.Type = "admin_api.delete_identity.executed"
)

type AdminAPIDeleteIdentityExecutedEventPayload struct {
	UserRef   model.UserRef  `json:"-" resolve:"user"`
	UserModel model.User     `json:"user"`
	Identity  model.Identity `json:"identity"`
}

func (e *AdminAPIDeleteIdentityExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIDeleteIdentityExecuted
}

func (e *AdminAPIDeleteIdentityExecutedEventPayload) UserID() string {
	return e.UserModel.ID
}

func (e *AdminAPIDeleteIdentityExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIDeleteIdentityExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPIDeleteIdentityExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIDeleteIdentityExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPIDeleteIdentityExecutedEventPayload) ReindexUserNeeded() bool {
	return false
}

func (e *AdminAPIDeleteIdentityExecutedEventPayload) IsUserDeleted() bool {
	return false
}

var _ event.NonBlockingPayload = &AdminAPIDeleteIdentityExecutedEventPayload{}

package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AdminAPICreateIdentityExecuted event.Type = "admin_api.create_identity.executed"
)

type AdminAPICreateIdentityExecutedEventPayload struct {
	UserRef   model.UserRef  `json:"-" resolve:"user"`
	UserModel model.User     `json:"user"`
	Identity  model.Identity `json:"identity"`
}

func (e *AdminAPICreateIdentityExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPICreateIdentityExecuted
}

func (e *AdminAPICreateIdentityExecutedEventPayload) UserID() string {
	return e.UserModel.ID
}

func (e *AdminAPICreateIdentityExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPICreateIdentityExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPICreateIdentityExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPICreateIdentityExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPICreateIdentityExecutedEventPayload) ReindexUserNeeded() bool {
	return false
}

func (e *AdminAPICreateIdentityExecutedEventPayload) IsUserDeleted() bool {
	return false
}

var _ event.NonBlockingPayload = &AdminAPICreateIdentityExecutedEventPayload{}

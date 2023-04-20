package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AdminAPIUpdateIdentityExecuted event.Type = "admin_api.update_identity.executed"
)

type AdminAPIUpdateIdentityExecutedEventPayload struct {
	UserRef   model.UserRef  `json:"-" resolve:"user"`
	UserModel model.User     `json:"user"`
	Identity  model.Identity `json:"identity"`
}

func (e *AdminAPIUpdateIdentityExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIUpdateIdentityExecuted
}

func (e *AdminAPIUpdateIdentityExecutedEventPayload) UserID() string {
	return e.UserModel.ID
}

func (e *AdminAPIUpdateIdentityExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIUpdateIdentityExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPIUpdateIdentityExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIUpdateIdentityExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPIUpdateIdentityExecutedEventPayload) ReindexUserNeeded() bool {
	return false
}

func (e *AdminAPIUpdateIdentityExecutedEventPayload) IsUserDeleted() bool {
	return false
}

var _ event.NonBlockingPayload = &AdminAPIUpdateIdentityExecutedEventPayload{}

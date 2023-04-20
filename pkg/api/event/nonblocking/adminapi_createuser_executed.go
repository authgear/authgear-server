package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AdminAPICreateUserExecuted event.Type = "admin_api.create_user.executed"
)

type AdminAPICreateUserExecutedEventPayload struct {
	UserRef   model.UserRef `json:"-" resolve:"user"`
	UserModel model.User    `json:"user"`
}

func (e *AdminAPICreateUserExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPICreateUserExecuted
}

func (e *AdminAPICreateUserExecutedEventPayload) UserID() string {
	return e.UserModel.ID
}

func (e *AdminAPICreateUserExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPICreateUserExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPICreateUserExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPICreateUserExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPICreateUserExecutedEventPayload) ReindexUserNeeded() bool {
	return false
}

func (e *AdminAPICreateUserExecutedEventPayload) IsUserDeleted() bool {
	return false
}

var _ event.NonBlockingPayload = &AdminAPICreateUserExecutedEventPayload{}

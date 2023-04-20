package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AdminAPIDeleteUserExecuted event.Type = "admin_api.delete_user.executed"
)

type AdminAPIDeleteUserExecutedEventPayload struct {
	UserModel model.User `json:"user"`
}

func (e *AdminAPIDeleteUserExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIDeleteUserExecuted
}

func (e *AdminAPIDeleteUserExecutedEventPayload) UserID() string {
	return e.UserModel.ID
}

func (e *AdminAPIDeleteUserExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIDeleteUserExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPIDeleteUserExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIDeleteUserExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPIDeleteUserExecutedEventPayload) ReindexUserNeeded() bool {
	return false
}

func (e *AdminAPIDeleteUserExecutedEventPayload) IsUserDeleted() bool {
	return false
}

var _ event.NonBlockingPayload = &AdminAPIDeleteUserExecutedEventPayload{}

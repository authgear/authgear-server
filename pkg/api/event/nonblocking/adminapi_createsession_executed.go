package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AdminAPICreateSessionExecuted event.Type = "admin_api.create_session.executed"
)

type AdminAPICreateSessionExecutedEventPayload struct {
	UserRef   model.UserRef `json:"-" resolve:"user"`
	UserModel model.User    `json:"user"`
}

func (e *AdminAPICreateSessionExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPICreateSessionExecuted
}

func (e *AdminAPICreateSessionExecutedEventPayload) UserID() string {
	return e.UserModel.ID
}

func (e *AdminAPICreateSessionExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPICreateSessionExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPICreateSessionExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPICreateSessionExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPICreateSessionExecutedEventPayload) ReindexUserNeeded() bool {
	return false
}

func (e *AdminAPICreateSessionExecutedEventPayload) IsUserDeleted() bool {
	return false
}

var _ event.NonBlockingPayload = &AdminAPICreateSessionExecutedEventPayload{}

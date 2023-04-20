package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AdminAPIUpdateUserExecuted event.Type = "admin_api.update_user.executed"
)

type AdminAPIUpdateUserExecutedEventPayload struct {
	UserRef   model.UserRef `json:"-" resolve:"user"`
	UserModel model.User    `json:"user"`
}

func (e *AdminAPIUpdateUserExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIUpdateUserExecuted
}

func (e *AdminAPIUpdateUserExecutedEventPayload) UserID() string {
	return e.UserModel.ID
}

func (e *AdminAPIUpdateUserExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIUpdateUserExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPIUpdateUserExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIUpdateUserExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPIUpdateUserExecutedEventPayload) ReindexUserNeeded() bool {
	return false
}

func (e *AdminAPIUpdateUserExecutedEventPayload) IsUserDeleted() bool {
	return false
}

var _ event.NonBlockingPayload = &AdminAPIUpdateUserExecutedEventPayload{}

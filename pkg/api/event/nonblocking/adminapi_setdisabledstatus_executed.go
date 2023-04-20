package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AdminAPISetDisabledStatusExecuted event.Type = "admin_api.set_disabled_status.executed"
)

type AdminAPISetDisabledStatusExecutedEventPayload struct {
	UserRef    model.UserRef `json:"-" resolve:"user"`
	UserModel  model.User    `json:"user"`
	IsDisabled bool          `json:"is_disabled"`
}

func (e *AdminAPISetDisabledStatusExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPISetDisabledStatusExecuted
}

func (e *AdminAPISetDisabledStatusExecutedEventPayload) UserID() string {
	return e.UserModel.ID
}

func (e *AdminAPISetDisabledStatusExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPISetDisabledStatusExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPISetDisabledStatusExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPISetDisabledStatusExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPISetDisabledStatusExecutedEventPayload) ReindexUserNeeded() bool {
	return false
}

func (e *AdminAPISetDisabledStatusExecutedEventPayload) IsUserDeleted() bool {
	return false
}

var _ event.NonBlockingPayload = &AdminAPISetDisabledStatusExecutedEventPayload{}

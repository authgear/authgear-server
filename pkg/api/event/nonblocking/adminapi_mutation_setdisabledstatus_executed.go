package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AdminAPIMutationSetDisabledStatusExecuted event.Type = "admin_api.mutation.set_disabled_status.executed"
)

type AdminAPIMutationSetDisabledStatusExecutedEventPayload struct {
	UserRef    model.UserRef `json:"-" resolve:"user"`
	UserModel  model.User    `json:"user"`
	IsDisabled bool          `json:"is_disabled"`
}

func (e *AdminAPIMutationSetDisabledStatusExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIMutationSetDisabledStatusExecuted
}

func (e *AdminAPIMutationSetDisabledStatusExecutedEventPayload) UserID() string {
	return e.UserModel.ID
}

func (e *AdminAPIMutationSetDisabledStatusExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIMutationSetDisabledStatusExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPIMutationSetDisabledStatusExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIMutationSetDisabledStatusExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPIMutationSetDisabledStatusExecutedEventPayload) ReindexUserNeeded() bool {
	return false
}

func (e *AdminAPIMutationSetDisabledStatusExecutedEventPayload) IsUserDeleted() bool {
	return false
}

var _ event.NonBlockingPayload = &AdminAPIMutationSetDisabledStatusExecutedEventPayload{}

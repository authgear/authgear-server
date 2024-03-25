package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AdminAPIMutationRevokeSessionExecuted event.Type = "admin_api.mutation.revoke_session.executed"
)

type AdminAPIMutationRevokeSessionExecutedEventPayload struct {
	UserRef   model.UserRef `json:"-" resolve:"user"`
	UserModel model.User    `json:"user"`
	Session   model.Session `json:"session"`
}

func (e *AdminAPIMutationRevokeSessionExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIMutationRevokeSessionExecuted
}

func (e *AdminAPIMutationRevokeSessionExecutedEventPayload) UserID() string {
	return e.UserModel.ID
}

func (e *AdminAPIMutationRevokeSessionExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIMutationRevokeSessionExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPIMutationRevokeSessionExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIMutationRevokeSessionExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPIMutationRevokeSessionExecutedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *AdminAPIMutationRevokeSessionExecutedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &AdminAPIMutationRevokeSessionExecutedEventPayload{}

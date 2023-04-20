package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AdminAPIRevokeSessionExecuted event.Type = "admin_api.revoke_session.executed"
)

type AdminAPIRevokeSessionExecutedEventPayload struct {
	UserRef   model.UserRef `json:"-" resolve:"user"`
	UserModel model.User    `json:"user"`
	SessionID string        `json:"session_id"`
}

func (e *AdminAPIRevokeSessionExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIRevokeSessionExecuted
}

func (e *AdminAPIRevokeSessionExecutedEventPayload) UserID() string {
	return e.UserModel.ID
}

func (e *AdminAPIRevokeSessionExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIRevokeSessionExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPIRevokeSessionExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIRevokeSessionExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPIRevokeSessionExecutedEventPayload) ReindexUserNeeded() bool {
	return false
}

func (e *AdminAPIRevokeSessionExecutedEventPayload) IsUserDeleted() bool {
	return false
}

var _ event.NonBlockingPayload = &AdminAPIRevokeSessionExecutedEventPayload{}

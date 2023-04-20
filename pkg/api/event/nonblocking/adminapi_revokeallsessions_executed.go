package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AdminAPIRevokeAllSessionsExecuted event.Type = "admin_api.revoke_all_sessions.executed"
)

type AdminAPIRevokeAllSessionsExecutedEventPayload struct {
	UserRef   model.UserRef `json:"-" resolve:"user"`
	UserModel model.User    `json:"user"`
}

func (e *AdminAPIRevokeAllSessionsExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIRevokeAllSessionsExecuted
}

func (e *AdminAPIRevokeAllSessionsExecutedEventPayload) UserID() string {
	return e.UserModel.ID
}

func (e *AdminAPIRevokeAllSessionsExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIRevokeAllSessionsExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPIRevokeAllSessionsExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIRevokeAllSessionsExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPIRevokeAllSessionsExecutedEventPayload) ReindexUserNeeded() bool {
	return false
}

func (e *AdminAPIRevokeAllSessionsExecutedEventPayload) IsUserDeleted() bool {
	return false
}

var _ event.NonBlockingPayload = &AdminAPIRevokeAllSessionsExecutedEventPayload{}

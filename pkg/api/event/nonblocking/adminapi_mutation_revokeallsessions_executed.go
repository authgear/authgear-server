package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AdminAPIMutationRevokeAllSessionsExecuted event.Type = "admin_api.mutation.revoke_all_sessions.executed"
)

type AdminAPIMutationRevokeAllSessionsExecutedEventPayload struct {
	UserRef   model.UserRef `json:"-" resolve:"user"`
	UserModel model.User    `json:"user"`
}

func (e *AdminAPIMutationRevokeAllSessionsExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIMutationRevokeAllSessionsExecuted
}

func (e *AdminAPIMutationRevokeAllSessionsExecutedEventPayload) UserID() string {
	return e.UserModel.ID
}

func (e *AdminAPIMutationRevokeAllSessionsExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIMutationRevokeAllSessionsExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPIMutationRevokeAllSessionsExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIMutationRevokeAllSessionsExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPIMutationRevokeAllSessionsExecutedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *AdminAPIMutationRevokeAllSessionsExecutedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &AdminAPIMutationRevokeAllSessionsExecutedEventPayload{}

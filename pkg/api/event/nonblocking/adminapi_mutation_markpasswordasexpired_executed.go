package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AdminAPIMutationMarkPasswordAsExpiredExecuted event.Type = "admin_api.mutation.mark_password_as_expired.executed" // nolint:gosec
)

type AdminAPIMutationMarkPasswordAsExpiredExecutedEventPayload struct {
	UserRef   model.UserRef `json:"-" resolve:"user"`
	UserModel model.User    `json:"user"`
	IsExpired bool          `json:"is_expired"`
}

func (e *AdminAPIMutationMarkPasswordAsExpiredExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIMutationMarkPasswordAsExpiredExecuted
}

func (e *AdminAPIMutationMarkPasswordAsExpiredExecutedEventPayload) UserID() string {
	return e.UserModel.ID
}

func (e *AdminAPIMutationMarkPasswordAsExpiredExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIMutationMarkPasswordAsExpiredExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPIMutationMarkPasswordAsExpiredExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIMutationMarkPasswordAsExpiredExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPIMutationMarkPasswordAsExpiredExecutedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *AdminAPIMutationMarkPasswordAsExpiredExecutedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &AdminAPIMutationMarkPasswordAsExpiredExecutedEventPayload{}

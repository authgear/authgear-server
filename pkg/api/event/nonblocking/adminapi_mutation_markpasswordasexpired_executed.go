package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AdminAPIMutationSetPasswordExpiredExecuted event.Type = "admin_api.mutation.set_password_expired.executed" // nolint:gosec
)

type AdminAPIMutationSetPasswordExpiredExecutedEventPayload struct {
	UserRef   model.UserRef `json:"-" resolve:"user"`
	UserModel model.User    `json:"user"`
	IsExpired bool          `json:"is_expired"`
}

func (e *AdminAPIMutationSetPasswordExpiredExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIMutationSetPasswordExpiredExecuted
}

func (e *AdminAPIMutationSetPasswordExpiredExecutedEventPayload) UserID() string {
	return e.UserModel.ID
}

func (e *AdminAPIMutationSetPasswordExpiredExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIMutationSetPasswordExpiredExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPIMutationSetPasswordExpiredExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIMutationSetPasswordExpiredExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPIMutationSetPasswordExpiredExecutedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *AdminAPIMutationSetPasswordExpiredExecutedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &AdminAPIMutationSetPasswordExpiredExecutedEventPayload{}

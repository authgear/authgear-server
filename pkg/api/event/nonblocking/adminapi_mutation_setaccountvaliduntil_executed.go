package nonblocking

import (
	"time"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AdminAPIMutationSetAccountValidUntilExecuted event.Type = "admin_api.mutation.set_account_valid_until.executed"
)

type AdminAPIMutationSetAccountValidUntilExecutedEventPayload struct {
	UserRef           model.UserRef `json:"-" resolve:"user"`
	UserModel         model.User    `json:"user"`
	AccountValidUntil *time.Time    `json:"account_valid_until"`
}

func (e *AdminAPIMutationSetAccountValidUntilExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIMutationSetAccountValidUntilExecuted
}

func (e *AdminAPIMutationSetAccountValidUntilExecutedEventPayload) UserID() string {
	return e.UserModel.ID
}

func (e *AdminAPIMutationSetAccountValidUntilExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIMutationSetAccountValidUntilExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPIMutationSetAccountValidUntilExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIMutationSetAccountValidUntilExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPIMutationSetAccountValidUntilExecutedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *AdminAPIMutationSetAccountValidUntilExecutedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &AdminAPIMutationSetAccountValidUntilExecutedEventPayload{}

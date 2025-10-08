package nonblocking

import (
	"time"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AdminAPIMutationSetAccountValidPeriodExecuted event.Type = "admin_api.mutation.set_account_valid_period.executed"
)

type AdminAPIMutationSetAccountValidPeriodExecutedEventPayload struct {
	UserRef           model.UserRef `json:"-" resolve:"user"`
	UserModel         model.User    `json:"user"`
	AccountValidFrom  *time.Time    `json:"account_valid_from"`
	AccountValidUntil *time.Time    `json:"account_valid_until"`
}

func (e *AdminAPIMutationSetAccountValidPeriodExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIMutationSetAccountValidPeriodExecuted
}

func (e *AdminAPIMutationSetAccountValidPeriodExecutedEventPayload) UserID() string {
	return e.UserModel.ID
}

func (e *AdminAPIMutationSetAccountValidPeriodExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIMutationSetAccountValidPeriodExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPIMutationSetAccountValidPeriodExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIMutationSetAccountValidPeriodExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPIMutationSetAccountValidPeriodExecutedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *AdminAPIMutationSetAccountValidPeriodExecutedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &AdminAPIMutationSetAccountValidPeriodExecutedEventPayload{}

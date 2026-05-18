package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AdminAPIMutationResetAccountLockoutExecuted event.Type = "admin_api.mutation.reset_account_lockout.executed"
)

type AdminAPIMutationResetAccountLockoutExecutedEventPayload struct {
	UserRef               model.UserRef               `json:"-" resolve:"user"`
	UserModel             model.User                  `json:"user"`
	PreviousLockoutStatus *model.AccountLockoutStatus `json:"previous_lockout_status"`
}

func (e *AdminAPIMutationResetAccountLockoutExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIMutationResetAccountLockoutExecuted
}

func (e *AdminAPIMutationResetAccountLockoutExecutedEventPayload) UserID() string {
	return e.UserModel.ID
}

func (e *AdminAPIMutationResetAccountLockoutExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIMutationResetAccountLockoutExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPIMutationResetAccountLockoutExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIMutationResetAccountLockoutExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPIMutationResetAccountLockoutExecutedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *AdminAPIMutationResetAccountLockoutExecutedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &AdminAPIMutationResetAccountLockoutExecutedEventPayload{}

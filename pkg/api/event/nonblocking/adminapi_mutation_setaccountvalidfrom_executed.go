package nonblocking

import (
	"time"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AdminAPIMutationSetAccountValidFromExecuted event.Type = "admin_api.mutation.set_account_valid_from.executed"
)

type AdminAPIMutationSetAccountValidFromExecutedEventPayload struct {
	UserRef          model.UserRef `json:"-" resolve:"user"`
	UserModel        model.User    `json:"user"`
	AccountValidFrom *time.Time    `json:"account_valid_from"`
}

func (e *AdminAPIMutationSetAccountValidFromExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIMutationSetAccountValidFromExecuted
}

func (e *AdminAPIMutationSetAccountValidFromExecutedEventPayload) UserID() string {
	return e.UserModel.ID
}

func (e *AdminAPIMutationSetAccountValidFromExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIMutationSetAccountValidFromExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPIMutationSetAccountValidFromExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIMutationSetAccountValidFromExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPIMutationSetAccountValidFromExecutedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *AdminAPIMutationSetAccountValidFromExecutedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &AdminAPIMutationSetAccountValidFromExecutedEventPayload{}

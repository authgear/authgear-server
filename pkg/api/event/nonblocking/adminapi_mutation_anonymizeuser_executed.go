package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AdminAPIMutationAnonymizeUserExecuted event.Type = "admin_api.mutation.anonymize_user.executed"
)

type AdminAPIMutationAnonymizeUserExecutedEventPayload struct {
	UserRef   model.UserRef `json:"-" resolve:"user"`
	UserModel model.User    `json:"user"`
}

func (e *AdminAPIMutationAnonymizeUserExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIMutationAnonymizeUserExecuted
}

func (e *AdminAPIMutationAnonymizeUserExecutedEventPayload) UserID() string {
	return e.UserModel.ID
}

func (e *AdminAPIMutationAnonymizeUserExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIMutationAnonymizeUserExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPIMutationAnonymizeUserExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIMutationAnonymizeUserExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPIMutationAnonymizeUserExecutedEventPayload) RequireReindexUserIDs() []string {
	return []string{e.UserID()}
}

func (e *AdminAPIMutationAnonymizeUserExecutedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &AdminAPIMutationAnonymizeUserExecutedEventPayload{}

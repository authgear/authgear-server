package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AdminAPIMutationDeleteAuthenticatorExecuted event.Type = "admin_api.mutation.delete_authenticator.executed"
)

type AdminAPIMutationDeleteAuthenticatorExecutedEventPayload struct {
	UserRef       model.UserRef       `json:"-" resolve:"user"`
	UserModel     model.User          `json:"user"`
	Authenticator model.Authenticator `json:"authenticator"`
}

func (e *AdminAPIMutationDeleteAuthenticatorExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIMutationDeleteAuthenticatorExecuted
}

func (e *AdminAPIMutationDeleteAuthenticatorExecutedEventPayload) UserID() string {
	return e.UserModel.ID
}

func (e *AdminAPIMutationDeleteAuthenticatorExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIMutationDeleteAuthenticatorExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPIMutationDeleteAuthenticatorExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIMutationDeleteAuthenticatorExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPIMutationDeleteAuthenticatorExecutedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *AdminAPIMutationDeleteAuthenticatorExecutedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &AdminAPIMutationDeleteAuthenticatorExecutedEventPayload{}

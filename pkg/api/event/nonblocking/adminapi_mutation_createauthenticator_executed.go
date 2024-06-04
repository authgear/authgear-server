package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AdminAPIMutationCreateAuthenticatorExecuted event.Type = "admin_api.mutation.create_authenticator.executed"
)

type AdminAPIMutationCreateAuthenticatorExecutedEventPayload struct {
	UserRef       model.UserRef       `json:"-" resolve:"user"`
	UserModel     model.User          `json:"user"`
	Authenticator model.Authenticator `json:"authenticator"`
}

func (e *AdminAPIMutationCreateAuthenticatorExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIMutationCreateAuthenticatorExecuted
}

func (e *AdminAPIMutationCreateAuthenticatorExecutedEventPayload) UserID() string {
	return e.UserModel.ID
}

func (e *AdminAPIMutationCreateAuthenticatorExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIMutationCreateAuthenticatorExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPIMutationCreateAuthenticatorExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIMutationCreateAuthenticatorExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPIMutationCreateAuthenticatorExecutedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *AdminAPIMutationCreateAuthenticatorExecutedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &AdminAPIMutationCreateAuthenticatorExecutedEventPayload{}

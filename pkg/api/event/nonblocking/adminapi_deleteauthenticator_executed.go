package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AdminAPIDeleteAuthenticatorExecuted event.Type = "admin_api.delete_authenticator.executed"
)

type AdminAPIDeleteAuthenticatorExecutedEventPayload struct {
	UserRef       model.UserRef       `json:"-" resolve:"user"`
	UserModel     model.User          `json:"user"`
	Authenticator model.Authenticator `json:"authenticator"`
}

func (e *AdminAPIDeleteAuthenticatorExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIDeleteAuthenticatorExecuted
}

func (e *AdminAPIDeleteAuthenticatorExecutedEventPayload) UserID() string {
	return e.UserModel.ID
}

func (e *AdminAPIDeleteAuthenticatorExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIDeleteAuthenticatorExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPIDeleteAuthenticatorExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIDeleteAuthenticatorExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPIDeleteAuthenticatorExecutedEventPayload) ReindexUserNeeded() bool {
	return false
}

func (e *AdminAPIDeleteAuthenticatorExecutedEventPayload) IsUserDeleted() bool {
	return false
}

var _ event.NonBlockingPayload = &AdminAPIDeleteAuthenticatorExecutedEventPayload{}

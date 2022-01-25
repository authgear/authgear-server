package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
)

const (
	AuthenticationFailedLoginID event.Type = "authentication.identity.login_id.failed"
)

type AuthenticationFailedLoginIDEventPayload struct {
	LoginID string `json:"login_id"`
}

func (e *AuthenticationFailedLoginIDEventPayload) NonBlockingEventType() event.Type {
	return AuthenticationFailedLoginID
}

func (e *AuthenticationFailedLoginIDEventPayload) UserID() string {
	return ""
}

func (e *AuthenticationFailedLoginIDEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeUser
}

func (e *AuthenticationFailedLoginIDEventPayload) FillContext(ctx *event.Context) {
	userID := e.UserID()
	ctx.UserID = &userID
}

func (e *AuthenticationFailedLoginIDEventPayload) ForWebHook() bool {
	return false
}

func (e *AuthenticationFailedLoginIDEventPayload) ForAudit() bool {
	return true
}

var _ event.NonBlockingPayload = &AuthenticationFailedLoginIDEventPayload{}

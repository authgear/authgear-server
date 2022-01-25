package nonblocking

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AuthenticationFailedIdentityFormat string = "authentication.identity.%s.failed"
)

type AuthenticationFailedIdentityEventPayload struct {
	UserRef      model.UserRef `json:"-" resolve:"user"`
	UserModel    model.User    `json:"user"`
	IdentityType string        `json:"identity_type"`
}

func (e *AuthenticationFailedIdentityEventPayload) NonBlockingEventType() event.Type {
	return event.Type(fmt.Sprintf(AuthenticationFailedIdentityFormat, e.IdentityType))
}

func (e *AuthenticationFailedIdentityEventPayload) UserID() string {
	return e.UserRef.ID
}

func (e *AuthenticationFailedIdentityEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeUser
}

func (e *AuthenticationFailedIdentityEventPayload) FillContext(ctx *event.Context) {
	userID := e.UserID()
	ctx.UserID = &userID
}

func (e *AuthenticationFailedIdentityEventPayload) ForWebHook() bool {
	return false
}

func (e *AuthenticationFailedIdentityEventPayload) ForAudit() bool {
	return true
}

var _ event.NonBlockingPayload = &AuthenticationFailedIdentityEventPayload{}

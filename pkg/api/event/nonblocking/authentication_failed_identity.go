package nonblocking

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AuthenticationFailedIdentityFormat string = "authentication.failed.identity.%s"
)

type AuthenticationFailedIdentityEventPayload struct {
	User         model.User `json:"user"`
	IdentityType string     `json:"identity_type"`
}

func (e *AuthenticationFailedIdentityEventPayload) NonBlockingEventType() event.Type {
	return event.Type(fmt.Sprintf(AuthenticationFailedIdentityFormat, e.IdentityType))
}

func (e *AuthenticationFailedIdentityEventPayload) UserID() string {
	return e.User.ID
}

func (e *AuthenticationFailedIdentityEventPayload) IsAdminAPI() bool {
	return false
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

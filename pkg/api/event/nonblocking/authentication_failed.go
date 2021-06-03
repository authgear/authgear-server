package nonblocking

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AuthenticationFailedFormat string = "authentication.%s.%s.failed"
)

type AuthenticationFailedEventPayload struct {
	User                model.User `json:"user"`
	AuthenticationStage string     `json:"authentication_stage"`
	AuthenticationType  string     `json:"authenticator_type"`
}

func (e *AuthenticationFailedEventPayload) NonBlockingEventType() event.Type {
	return event.Type(fmt.Sprintf(AuthenticationFailedFormat, e.AuthenticationStage, e.AuthenticationType))
}

func (e *AuthenticationFailedEventPayload) UserID() string {
	return e.User.ID
}

func (e *AuthenticationFailedEventPayload) IsAdminAPI() bool {
	return false
}

func (e *AuthenticationFailedEventPayload) FillContext(ctx *event.Context) {
	userID := e.UserID()
	ctx.UserID = &userID
}

func (e *AuthenticationFailedEventPayload) ForWebHook() bool {
	return false
}

func (e *AuthenticationFailedEventPayload) ForAudit() bool {
	return true
}

var _ event.NonBlockingPayload = &AuthenticationFailedEventPayload{}

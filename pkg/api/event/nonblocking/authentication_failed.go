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
	UserRef             model.UserRef `json:"-" resolve:"user"`
	UserModel           model.User    `json:"user"`
	AuthenticationStage string        `json:"authentication_stage"`
	AuthenticationType  string        `json:"authenticator_type"`
}

func (e *AuthenticationFailedEventPayload) NonBlockingEventType() event.Type {
	return event.Type(fmt.Sprintf(AuthenticationFailedFormat, e.AuthenticationStage, e.AuthenticationType))
}

func (e *AuthenticationFailedEventPayload) UserID() string {
	return e.UserRef.ID
}

func (e *AuthenticationFailedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeUser
}

func (e *AuthenticationFailedEventPayload) FillContext(ctx *event.Context) {
	userID := e.UserID()
	ctx.UserID = &userID
}

func (e *AuthenticationFailedEventPayload) ForHook() bool {
	return false
}

func (e *AuthenticationFailedEventPayload) ForAudit() bool {
	return true
}

func (e *AuthenticationFailedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *AuthenticationFailedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &AuthenticationFailedEventPayload{}

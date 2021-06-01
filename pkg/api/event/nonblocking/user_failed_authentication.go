package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	UserFailedAuthentication event.Type = "user.failed_authentication"
)

type UserFailedAuthenticationEventPayload struct {
	User model.User `json:"user"`
}

func (e *UserFailedAuthenticationEventPayload) NonBlockingEventType() event.Type {
	return UserFailedAuthentication
}

func (e *UserFailedAuthenticationEventPayload) UserID() string {
	return e.User.ID
}

func (e *UserFailedAuthenticationEventPayload) IsAdminAPI() bool {
	return false
}

func (e *UserFailedAuthenticationEventPayload) FillContext(ctx *event.Context) {
	userID := e.UserID()
	ctx.UserID = &userID
}

var _ event.NonBlockingPayload = &UserFailedAuthenticationEventPayload{}

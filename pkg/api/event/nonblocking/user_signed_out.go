package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	UserSignedOut event.Type = "user.signed_out"
)

type UserSignedOutEventPayload struct {
	User     model.User    `json:"user"`
	Session  model.Session `json:"session"`
	AdminAPI bool          `json:"-"`
}

func (e *UserSignedOutEventPayload) NonBlockingEventType() event.Type {
	return UserSignedOut
}

func (e *UserSignedOutEventPayload) UserID() string {
	return e.User.ID
}

func (e *UserSignedOutEventPayload) IsAdminAPI() bool {
	return e.AdminAPI
}

func (e *UserSignedOutEventPayload) FillContext(ctx *event.Context) {
	userID := e.UserID()
	ctx.UserID = &userID
}

var _ event.NonBlockingPayload = &UserSignedOutEventPayload{}

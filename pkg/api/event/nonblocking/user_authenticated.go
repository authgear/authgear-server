package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	UserAuthenticated event.Type = "user.authenticated"
)

type UserAuthenticatedEventPayload struct {
	User     model.User    `json:"user"`
	Session  model.Session `json:"session"`
	AdminAPI bool          `json:"-"`
}

func (e *UserAuthenticatedEventPayload) NonBlockingEventType() event.Type {
	return UserAuthenticated
}

func (e *UserAuthenticatedEventPayload) UserID() string {
	return e.User.ID
}

func (e *UserAuthenticatedEventPayload) IsAdminAPI() bool {
	return e.AdminAPI
}

func (e *UserAuthenticatedEventPayload) FillContext(ctx *event.Context) {
}

func (e *UserAuthenticatedEventPayload) ForWebHook() bool {
	return true
}

func (e *UserAuthenticatedEventPayload) ForAudit() bool {
	return true
}

var _ event.NonBlockingPayload = &UserAuthenticatedEventPayload{}

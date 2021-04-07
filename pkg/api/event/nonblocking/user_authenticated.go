package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	UserAuthenticated event.Type = "user.authenticated"
)

type UserAuthenticatedEvent struct {
	User     model.User    `json:"user"`
	Session  model.Session `json:"session"`
	AdminAPI bool          `json:"-"`
}

func (e *UserAuthenticatedEvent) NonBlockingEventType() event.Type {
	return UserAuthenticated
}

func (e *UserAuthenticatedEvent) UserID() string {
	return e.User.ID
}

func (e *UserAuthenticatedEvent) IsAdminAPI() bool {
	return e.AdminAPI
}

var _ event.NonBlockingPayload = &UserAuthenticatedEvent{}

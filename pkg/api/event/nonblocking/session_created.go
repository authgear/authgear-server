package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	SessionCreatedUserSignup            event.Type = "session.created.user_signup"
	SessionCreatedUserLogin             event.Type = "session.created.user_login"
	SessionCreatedUserPromoteThemselves event.Type = "session.created.user_promote_themselves"
)

type SessionCreatedUserSignupEvent struct {
	User    model.User    `json:"user"`
	Session model.Session `json:"session"`
}

func (e *SessionCreatedUserSignupEvent) NonBlockingEventType() event.Type {
	return SessionCreatedUserSignup
}

func (e *SessionCreatedUserSignupEvent) UserID() string {
	return e.User.ID
}

type SessionCreatedUserLoginEvent struct {
	User    model.User    `json:"user"`
	Session model.Session `json:"session"`
}

func (e *SessionCreatedUserLoginEvent) NonBlockingEventType() event.Type {
	return SessionCreatedUserLogin
}

func (e *SessionCreatedUserLoginEvent) UserID() string {
	return e.User.ID
}

type SessionCreatedUserPromoteThemselvesEvent struct {
	User    model.User    `json:"user"`
	Session model.Session `json:"session"`
}

func (e *SessionCreatedUserPromoteThemselvesEvent) NonBlockingEventType() event.Type {
	return SessionCreatedUserPromoteThemselves
}

func (e *SessionCreatedUserPromoteThemselvesEvent) UserID() string {
	return e.User.ID
}

var _ event.NonBlockingPayload = &SessionCreatedUserSignupEvent{}
var _ event.NonBlockingPayload = &SessionCreatedUserLoginEvent{}
var _ event.NonBlockingPayload = &SessionCreatedUserPromoteThemselvesEvent{}

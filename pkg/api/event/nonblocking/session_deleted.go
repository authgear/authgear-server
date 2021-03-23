package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	SessionDeletedUserRevokeSession     event.Type = "session.deleted.user_revoke_session"
	SessionDeletedUserLogout            event.Type = "session.deleted.user_logout"
	SessionDeletedAdminAPIRevokeSession event.Type = "session.deleted.admin_api_revoke_session"
)

type SessionDeletedUserRevokeSessionEvent struct {
	User    model.User    `json:"user"`
	Session model.Session `json:"session"`
}

func (e *SessionDeletedUserRevokeSessionEvent) NonBlockingEventType() event.Type {
	return SessionDeletedUserRevokeSession
}

func (e *SessionDeletedUserRevokeSessionEvent) UserID() string {
	return e.User.ID
}

type SessionDeletedUserLogoutEvent struct {
	User    model.User    `json:"user"`
	Session model.Session `json:"session"`
}

func (e *SessionDeletedUserLogoutEvent) NonBlockingEventType() event.Type {
	return SessionDeletedUserLogout
}

func (e *SessionDeletedUserLogoutEvent) UserID() string {
	return e.User.ID
}

type SessionDeletedAdminAPIRevokeSessionEvent struct {
	User    model.User    `json:"user"`
	Session model.Session `json:"session"`
}

func (e *SessionDeletedAdminAPIRevokeSessionEvent) NonBlockingEventType() event.Type {
	return SessionDeletedAdminAPIRevokeSession
}

func (e *SessionDeletedAdminAPIRevokeSessionEvent) UserID() string {
	return e.User.ID
}

var _ event.NonBlockingPayload = &SessionDeletedUserRevokeSessionEvent{}
var _ event.NonBlockingPayload = &SessionDeletedUserLogoutEvent{}
var _ event.NonBlockingPayload = &SessionDeletedAdminAPIRevokeSessionEvent{}

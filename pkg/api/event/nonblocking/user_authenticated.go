package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	UserAuthenticated event.Type = "user.authenticated"
)

type UserAuthenticatedEventPayload struct {
	UserRef   model.UserRef `json:"-" resolve:"user"`
	UserModel model.User    `json:"user"`
	Session   model.Session `json:"session"`
	AdminAPI  bool          `json:"-"`
}

func (e *UserAuthenticatedEventPayload) NonBlockingEventType() event.Type {
	return UserAuthenticated
}

func (e *UserAuthenticatedEventPayload) UserID() string {
	return e.UserRef.ID
}

func (e *UserAuthenticatedEventPayload) GetTriggeredBy() event.TriggeredByType {
	if e.AdminAPI {
		return event.TriggeredByTypeAdminAPI
	}
	return event.TriggeredByTypeUser
}

func (e *UserAuthenticatedEventPayload) FillContext(ctx *event.Context) {
}

func (e *UserAuthenticatedEventPayload) ForHook() bool {
	return true
}

func (e *UserAuthenticatedEventPayload) ForAudit() bool {
	return true
}

func (e *UserAuthenticatedEventPayload) RequireReindexUserIDs() []string {
	return []string{e.UserID()}
}

func (e *UserAuthenticatedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &UserAuthenticatedEventPayload{}

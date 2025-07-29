package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	UserReauthenticated event.Type = "user.reauthenticated"
)

type UserReauthenticatedEventPayload struct {
	UserRef   model.UserRef `json:"-" resolve:"user"`
	UserModel model.User    `json:"user"`
	Session   model.Session `json:"session"`
	AdminAPI  bool          `json:"-"`
}

func (e *UserReauthenticatedEventPayload) NonBlockingEventType() event.Type {
	return UserReauthenticated
}

func (e *UserReauthenticatedEventPayload) UserID() string {
	return e.UserRef.ID
}

func (e *UserReauthenticatedEventPayload) GetTriggeredBy() event.TriggeredByType {
	if e.AdminAPI {
		return event.TriggeredByTypeAdminAPI
	}
	return event.TriggeredByTypeUser
}

func (e *UserReauthenticatedEventPayload) FillContext(ctx *event.Context) {
}

func (e *UserReauthenticatedEventPayload) ForHook() bool {
	return true
}

func (e *UserReauthenticatedEventPayload) ForAudit() bool {
	return true
}

func (e *UserReauthenticatedEventPayload) RequireReindexUserIDs() []string {
	return []string{e.UserID()}
}

func (e *UserReauthenticatedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &UserReauthenticatedEventPayload{}

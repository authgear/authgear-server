package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	UserAnonymousPromoted event.Type = "user.anonymous.promoted"
)

type UserAnonymousPromotedEventPayload struct {
	AnonymousUserRef   model.UserRef    `json:"-" resolve:"anonymous_user"`
	AnonymousUserModel model.User       `json:"anonymous_user"`
	UserRef            model.UserRef    `json:"-" resolve:"user"`
	UserModel          model.User       `json:"user"`
	Identities         []model.Identity `json:"identities"`
	AdminAPI           bool             `json:"-"`
}

func (e *UserAnonymousPromotedEventPayload) NonBlockingEventType() event.Type {
	return UserAnonymousPromoted
}

func (e *UserAnonymousPromotedEventPayload) UserID() string {
	return e.UserRef.ID
}

func (e *UserAnonymousPromotedEventPayload) GetTriggeredBy() event.TriggeredByType {
	if e.AdminAPI {
		return event.TriggeredByTypeAdminAPI
	}
	return event.TriggeredByTypeUser
}

func (e *UserAnonymousPromotedEventPayload) FillContext(ctx *event.Context) {
}

func (e *UserAnonymousPromotedEventPayload) ForHook() bool {
	return true
}

func (e *UserAnonymousPromotedEventPayload) ForAudit() bool {
	return true
}

func (e *UserAnonymousPromotedEventPayload) RequireReindexUserIDs() []string {
	return []string{e.UserID()}
}

func (e *UserAnonymousPromotedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &UserAnonymousPromotedEventPayload{}

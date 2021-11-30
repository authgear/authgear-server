package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	UserAnonymousPromoted event.Type = "user.anonymous.promoted"
)

type UserAnonymousPromotedEventPayload struct {
	AnonymousUserRef   model.UserRef    `json:"-"`
	AnonymousUserModel model.User       `json:"anonymous_user"`
	UserRef            model.UserRef    `json:"-"`
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

func (e *UserAnonymousPromotedEventPayload) IsAdminAPI() bool {
	return e.AdminAPI
}

func (e *UserAnonymousPromotedEventPayload) FillContext(ctx *event.Context) {
}

func (e *UserAnonymousPromotedEventPayload) ForWebHook() bool {
	return true
}

func (e *UserAnonymousPromotedEventPayload) ForAudit() bool {
	return true
}

var _ event.NonBlockingPayload = &UserAnonymousPromotedEventPayload{}

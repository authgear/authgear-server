package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	UserAnonymousPromoted event.Type = "user.anonymous.promoted"
)

type UserAnonymousPromotedEventPayload struct {
	AnonymousUser model.User       `json:"anonymous_user"`
	User          model.User       `json:"user"`
	Identities    []model.Identity `json:"identities"`
	AdminAPI      bool             `json:"-"`
}

func (e *UserAnonymousPromotedEventPayload) NonBlockingEventType() event.Type {
	return UserAnonymousPromoted
}

func (e *UserAnonymousPromotedEventPayload) UserID() string {
	return e.User.ID
}

func (e *UserAnonymousPromotedEventPayload) IsAdminAPI() bool {
	return e.AdminAPI
}

func (e *UserAnonymousPromotedEventPayload) FillContext(ctx *event.Context) {
}

var _ event.NonBlockingPayload = &UserAnonymousPromotedEventPayload{}

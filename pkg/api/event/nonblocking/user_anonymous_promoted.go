package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	UserAnonymousPromoted event.Type = "user.anonymous.promoted"
)

type UserAnonymousPromotedEvent struct {
	AnonymousUser model.User       `json:"anonymous_user"`
	User          model.User       `json:"user"`
	Identities    []model.Identity `json:"identities"`
}

func (e *UserAnonymousPromotedEvent) NonBlockingEventType() event.Type {
	return UserAnonymousPromoted
}

func (e *UserAnonymousPromotedEvent) UserID() string {
	return e.User.ID
}

var _ event.NonBlockingPayload = &UserAnonymousPromotedEvent{}

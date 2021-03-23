package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	UserPromoted event.Type = "user.promoted.user_promote_themselves"
)

type UserPromotedEvent struct {
	AnonymousUser model.User       `json:"anonymous_user"`
	User          model.User       `json:"user"`
	Identities    []model.Identity `json:"identities"`
}

func (e *UserPromotedEvent) NonBlockingEventType() event.Type {
	return UserPromoted
}

func (e *UserPromotedEvent) UserID() string {
	return e.User.ID
}

var _ event.NonBlockingPayload = &UserPromotedEvent{}

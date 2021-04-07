package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	UserCreated event.Type = "user.created"
)

type UserCreatedEvent struct {
	User       model.User       `json:"user"`
	Identities []model.Identity `json:"identities"`
	AdminAPI   bool             `json:"-"`
}

func (e *UserCreatedEvent) NonBlockingEventType() event.Type {
	return UserCreated
}

func (e *UserCreatedEvent) UserID() string {
	return e.User.ID
}

func (e *UserCreatedEvent) IsAdminAPI() bool {
	return e.AdminAPI
}

var _ event.NonBlockingPayload = &UserCreatedEvent{}

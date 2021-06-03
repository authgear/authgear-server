package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	UserCreated event.Type = "user.created"
)

type UserCreatedEventPayload struct {
	User       model.User       `json:"user"`
	Identities []model.Identity `json:"identities"`
	AdminAPI   bool             `json:"-"`
}

func (e *UserCreatedEventPayload) NonBlockingEventType() event.Type {
	return UserCreated
}

func (e *UserCreatedEventPayload) UserID() string {
	return e.User.ID
}

func (e *UserCreatedEventPayload) IsAdminAPI() bool {
	return e.AdminAPI
}

func (e *UserCreatedEventPayload) FillContext(ctx *event.Context) {
}

func (e *UserCreatedEventPayload) ForWebHook() bool {
	return true
}

func (e *UserCreatedEventPayload) ForAudit() bool {
	return true
}

var _ event.NonBlockingPayload = &UserCreatedEventPayload{}

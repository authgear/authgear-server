package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	UserDisabled event.Type = "user.disabled"
)

type UserDisabledEventPayload struct {
	UserRef   model.UserRef `json:"-" resolve:"user"`
	UserModel model.User    `json:"user"`
}

func (e *UserDisabledEventPayload) NonBlockingEventType() event.Type {
	return UserDisabled
}

func (e *UserDisabledEventPayload) UserID() string {
	return e.UserRef.ID
}

func (e *UserDisabledEventPayload) IsAdminAPI() bool {
	return true
}

func (e *UserDisabledEventPayload) FillContext(ctx *event.Context) {
}

func (e *UserDisabledEventPayload) ForWebHook() bool {
	return true
}

func (e *UserDisabledEventPayload) ForAudit() bool {
	return true
}

var _ event.NonBlockingPayload = &UserDisabledEventPayload{}

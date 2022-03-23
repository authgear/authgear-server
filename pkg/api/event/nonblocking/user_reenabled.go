package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	UserReenabled event.Type = "user.reenabled"
)

type UserReenabledEventPayload struct {
	UserRef   model.UserRef `json:"-" resolve:"user"`
	UserModel model.User    `json:"user"`
}

func (e *UserReenabledEventPayload) NonBlockingEventType() event.Type {
	return UserReenabled
}

func (e *UserReenabledEventPayload) UserID() string {
	return e.UserRef.ID
}

func (e *UserReenabledEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *UserReenabledEventPayload) FillContext(ctx *event.Context) {
}

func (e *UserReenabledEventPayload) ForWebHook() bool {
	return true
}

func (e *UserReenabledEventPayload) ForAudit() bool {
	return true
}

func (e *UserReenabledEventPayload) ReindexUserNeeded() bool {
	return true
}

func (e *UserReenabledEventPayload) IsUserDeleted() bool {
	return false
}

var _ event.NonBlockingPayload = &UserReenabledEventPayload{}

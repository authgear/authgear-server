package blocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	UserPreCreate event.Type = "user.pre_create"
)

type UserPreCreateBlockingEvent struct {
	User       model.User       `json:"user"`
	Identities []model.Identity `json:"identities"`
}

func (e *UserPreCreateBlockingEvent) BlockingEventType() event.Type {
	return UserPreCreate
}

func (e *UserPreCreateBlockingEvent) UserID() string {
	return e.User.ID
}

var _ event.BlockingPayload = &UserPreCreateBlockingEvent{}

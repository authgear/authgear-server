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
	State      string           `json:"state,omitempty"`
	AdminAPI   bool             `json:"-"`
}

func (e *UserPreCreateBlockingEvent) BlockingEventType() event.Type {
	return UserPreCreate
}

func (e *UserPreCreateBlockingEvent) UserID() string {
	return e.User.ID
}

func (e *UserPreCreateBlockingEvent) IsAdminAPI() bool {
	return e.AdminAPI
}

var _ event.BlockingPayload = &UserPreCreateBlockingEvent{}

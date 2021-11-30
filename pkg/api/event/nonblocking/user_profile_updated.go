package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	UserProfileUpdated event.Type = "user.profile.updated"
)

type UserProfileUpdatedEventPayload struct {
	UserRef   model.UserRef `json:"-"`
	UserModel model.User    `json:"user"`
	AdminAPI  bool          `json:"-"`
}

func (e *UserProfileUpdatedEventPayload) NonBlockingEventType() event.Type {
	return UserProfileUpdated
}

func (e *UserProfileUpdatedEventPayload) UserID() string {
	return e.UserRef.ID
}

func (e *UserProfileUpdatedEventPayload) IsAdminAPI() bool {
	return e.AdminAPI
}

func (e *UserProfileUpdatedEventPayload) FillContext(ctx *event.Context) {
}

func (e *UserProfileUpdatedEventPayload) ForWebHook() bool {
	return true
}

func (e *UserProfileUpdatedEventPayload) ForAudit() bool {
	return true
}

var _ event.NonBlockingPayload = &UserProfileUpdatedEventPayload{}

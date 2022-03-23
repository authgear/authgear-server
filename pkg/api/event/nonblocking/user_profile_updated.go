package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	UserProfileUpdated event.Type = "user.profile.updated"
)

type UserProfileUpdatedEventPayload struct {
	UserRef   model.UserRef `json:"-" resolve:"user"`
	UserModel model.User    `json:"user"`
	AdminAPI  bool          `json:"-"`
}

func (e *UserProfileUpdatedEventPayload) NonBlockingEventType() event.Type {
	return UserProfileUpdated
}

func (e *UserProfileUpdatedEventPayload) UserID() string {
	return e.UserRef.ID
}

func (e *UserProfileUpdatedEventPayload) GetTriggeredBy() event.TriggeredByType {
	if e.AdminAPI {
		return event.TriggeredByTypeAdminAPI
	}
	return event.TriggeredByTypeUser
}

func (e *UserProfileUpdatedEventPayload) FillContext(ctx *event.Context) {
}

func (e *UserProfileUpdatedEventPayload) ForWebHook() bool {
	return true
}

func (e *UserProfileUpdatedEventPayload) ForAudit() bool {
	return true
}

func (e *UserProfileUpdatedEventPayload) ReindexUserNeeded() bool {
	return true
}

func (e *UserProfileUpdatedEventPayload) IsUserDeleted() bool {
	return false
}

var _ event.NonBlockingPayload = &UserProfileUpdatedEventPayload{}

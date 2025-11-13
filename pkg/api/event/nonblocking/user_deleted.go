package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	UserDeleted event.Type = "user.deleted"
)

type UserDeletedEventPayload struct {
	// We cannot use UserRef here because the user will be deleted BEFORE retrieval.
	UserModel           model.User `json:"user"`
	Reason              string     `json:"reason,omitzero"`
	IsScheduledDeletion bool       `json:"-"`
}

func (e *UserDeletedEventPayload) NonBlockingEventType() event.Type {
	return UserDeleted
}

func (e *UserDeletedEventPayload) UserID() string {
	return e.UserModel.ID
}

func (e *UserDeletedEventPayload) GetTriggeredBy() event.TriggeredByType {
	if e.IsScheduledDeletion {
		return event.TriggeredBySystem
	}
	return event.TriggeredByTypeAdminAPI
}

func (e *UserDeletedEventPayload) FillContext(ctx *event.Context) {
}

func (e *UserDeletedEventPayload) ForHook() bool {
	return true
}

func (e *UserDeletedEventPayload) ForAudit() bool {
	return true
}

func (e *UserDeletedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *UserDeletedEventPayload) DeletedUserIDs() []string {
	return []string{e.UserID()}
}

var _ event.NonBlockingPayload = &UserDeletedEventPayload{}

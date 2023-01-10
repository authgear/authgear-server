package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	UserAnonymized event.Type = "user.anonymized"
)

type UserAnonymizedEventPayload struct {
	// We cannot use UserRef here because the user will be Anonymized BEFORE retrieval.
	UserModel                model.User `json:"user"`
	IsScheduledAnonymization bool       `json:"-"`
}

func (e *UserAnonymizedEventPayload) NonBlockingEventType() event.Type {
	return UserAnonymized
}

func (e *UserAnonymizedEventPayload) UserID() string {
	return e.UserModel.ID
}

func (e *UserAnonymizedEventPayload) GetTriggeredBy() event.TriggeredByType {
	if e.IsScheduledAnonymization {
		return event.TriggeredBySystem
	}
	return event.TriggeredByTypeAdminAPI
}

func (e *UserAnonymizedEventPayload) FillContext(ctx *event.Context) {
}

func (e *UserAnonymizedEventPayload) ForHook() bool {
	return true
}

func (e *UserAnonymizedEventPayload) ForAudit() bool {
	return true
}

func (e *UserAnonymizedEventPayload) ReindexUserNeeded() bool {
	return true
}

func (e *UserAnonymizedEventPayload) IsUserDeleted() bool {
	return false
}

var _ event.NonBlockingPayload = &UserAnonymizedEventPayload{}

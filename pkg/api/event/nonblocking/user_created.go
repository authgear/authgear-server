package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	UserCreated event.Type = "user.created"
)

type UserCreatedEventPayload struct {
	UserRef    model.UserRef    `json:"-" resolve:"user"`
	UserModel  model.User       `json:"user"`
	Identities []model.Identity `json:"identities"`
	AdminAPI   bool             `json:"-"`
}

func (e *UserCreatedEventPayload) NonBlockingEventType() event.Type {
	return UserCreated
}

func (e *UserCreatedEventPayload) UserID() string {
	return e.UserRef.ID
}

func (e *UserCreatedEventPayload) GetTriggeredBy() event.TriggeredByType {
	if e.AdminAPI {
		return event.TriggeredByTypeAdminAPI
	}
	return event.TriggeredByTypeUser
}

func (e *UserCreatedEventPayload) FillContext(ctx *event.Context) {
}

func (e *UserCreatedEventPayload) ForHook() bool {
	return true
}

func (e *UserCreatedEventPayload) ForAudit() bool {
	return true
}

func (e *UserCreatedEventPayload) RequireReindexUserIDs() []string {
	return []string{e.UserID()}
}

func (e *UserCreatedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &UserCreatedEventPayload{}

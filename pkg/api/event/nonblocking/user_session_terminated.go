package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

type UserSessionTerminationType string

const (
	UserSessionTerminationTypeIndividual       UserSessionTerminationType = "individual"
	UserSessionTerminationTypeAll              UserSessionTerminationType = "all"
	UserSessionTerminationTypeAllExceptCurrent UserSessionTerminationType = "all_except_current"
)

const (
	UserSessionTerminated event.Type = "user.session.terminated"
)

type UserSessionTerminatedEventPayload struct {
	UserRef         model.UserRef              `json:"-" resolve:"user"`
	UserModel       model.User                 `json:"user"`
	Sessions        []model.Session            `json:"sessions"`
	AdminAPI        bool                       `json:"-"`
	TerminationType UserSessionTerminationType `json:"termination_type"`
}

func (e *UserSessionTerminatedEventPayload) NonBlockingEventType() event.Type {
	return UserSessionTerminated
}

func (e *UserSessionTerminatedEventPayload) UserID() string {
	return e.UserRef.ID
}

func (e *UserSessionTerminatedEventPayload) GetTriggeredBy() event.TriggeredByType {
	if e.AdminAPI {
		return event.TriggeredByTypeAdminAPI
	}
	return event.TriggeredByTypeUser
}

func (e *UserSessionTerminatedEventPayload) FillContext(ctx *event.Context) {
}

func (e *UserSessionTerminatedEventPayload) ForHook() bool {
	return false
}

func (e *UserSessionTerminatedEventPayload) ForAudit() bool {
	return true
}

func (e *UserSessionTerminatedEventPayload) ReindexUserNeeded() bool {
	return false
}

func (e *UserSessionTerminatedEventPayload) IsUserDeleted() bool {
	return false
}

var _ event.NonBlockingPayload = &UserSessionTerminatedEventPayload{}

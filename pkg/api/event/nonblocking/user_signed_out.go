package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	UserSignedOut event.Type = "user.signed_out"
)

type UserSignedOutEventPayload struct {
	UserRef   model.UserRef `json:"-" resolve:"user"`
	UserModel model.User    `json:"user"`
	// Replaced by sessions
	// Session   model.Session `json:"session"`
	Sessions []model.Session `json:"sessions"`
	AdminAPI bool            `json:"-"`
}

func (e *UserSignedOutEventPayload) NonBlockingEventType() event.Type {
	return UserSignedOut
}

func (e *UserSignedOutEventPayload) UserID() string {
	return e.UserRef.ID
}

func (e *UserSignedOutEventPayload) GetTriggeredBy() event.TriggeredByType {
	if e.AdminAPI {
		return event.TriggeredByTypeAdminAPI
	}
	return event.TriggeredByTypeUser
}

func (e *UserSignedOutEventPayload) FillContext(ctx *event.Context) {
}

func (e *UserSignedOutEventPayload) ForHook() bool {
	return false
}

func (e *UserSignedOutEventPayload) ForAudit() bool {
	return true
}

func (e *UserSignedOutEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *UserSignedOutEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &UserSignedOutEventPayload{}

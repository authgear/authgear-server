package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	IdentityOAuthDisconnected event.Type = "identity.oauth.disconnected"
)

type IdentityOAuthDisconnectedEventPayload struct {
	UserRef   model.UserRef  `json:"-" resolve:"user"`
	UserModel model.User     `json:"user"`
	Identity  model.Identity `json:"identity"`
	AdminAPI  bool           `json:"-"`
}

func (e *IdentityOAuthDisconnectedEventPayload) NonBlockingEventType() event.Type {
	return IdentityOAuthDisconnected
}

func (e *IdentityOAuthDisconnectedEventPayload) UserID() string {
	return e.UserRef.ID
}

func (e *IdentityOAuthDisconnectedEventPayload) GetTriggeredBy() event.TriggeredByType {
	if e.AdminAPI {
		return event.TriggeredByTypeAdminAPI
	}
	return event.TriggeredByTypeUser
}

func (e *IdentityOAuthDisconnectedEventPayload) FillContext(ctx *event.Context) {
}

func (e *IdentityOAuthDisconnectedEventPayload) ForHook() bool {
	return true
}

func (e *IdentityOAuthDisconnectedEventPayload) ForAudit() bool {
	return true
}

func (e *IdentityOAuthDisconnectedEventPayload) RequireReindexUserIDs() []string {
	return []string{e.UserID()}
}

func (e *IdentityOAuthDisconnectedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &IdentityOAuthDisconnectedEventPayload{}

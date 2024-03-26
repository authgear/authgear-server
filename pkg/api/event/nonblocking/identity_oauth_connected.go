package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	IdentityOAuthConnected event.Type = "identity.oauth.connected"
)

type IdentityOAuthConnectedEventPayload struct {
	UserRef   model.UserRef  `json:"-" resolve:"user"`
	UserModel model.User     `json:"user"`
	Identity  model.Identity `json:"identity"`
	AdminAPI  bool           `json:"-"`
}

func (e *IdentityOAuthConnectedEventPayload) NonBlockingEventType() event.Type {
	return IdentityOAuthConnected
}

func (e *IdentityOAuthConnectedEventPayload) UserID() string {
	return e.UserRef.ID
}

func (e *IdentityOAuthConnectedEventPayload) GetTriggeredBy() event.TriggeredByType {
	if e.AdminAPI {
		return event.TriggeredByTypeAdminAPI
	}
	return event.TriggeredByTypeUser
}

func (e *IdentityOAuthConnectedEventPayload) FillContext(ctx *event.Context) {
}

func (e *IdentityOAuthConnectedEventPayload) ForHook() bool {
	return true
}

func (e *IdentityOAuthConnectedEventPayload) ForAudit() bool {
	return true
}

func (e *IdentityOAuthConnectedEventPayload) RequireReindexUserIDs() []string {
	return []string{e.UserID()}
}

func (e *IdentityOAuthConnectedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &IdentityOAuthConnectedEventPayload{}

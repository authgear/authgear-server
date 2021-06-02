package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	IdentityOAuthConnected event.Type = "identity.oauth.connected"
)

type IdentityOAuthConnectedEventPayload struct {
	User     model.User     `json:"user"`
	Identity model.Identity `json:"identity"`
	AdminAPI bool           `json:"-"`
}

func (e *IdentityOAuthConnectedEventPayload) NonBlockingEventType() event.Type {
	return IdentityOAuthConnected
}

func (e *IdentityOAuthConnectedEventPayload) UserID() string {
	return e.User.ID
}

func (e *IdentityOAuthConnectedEventPayload) IsAdminAPI() bool {
	return e.AdminAPI
}

func (e *IdentityOAuthConnectedEventPayload) FillContext(ctx *event.Context) {
}

func (e *IdentityOAuthConnectedEventPayload) ForWebHook() bool {
	return true
}

func (e *IdentityOAuthConnectedEventPayload) ForAudit() bool {
	return true
}

var _ event.NonBlockingPayload = &IdentityOAuthConnectedEventPayload{}

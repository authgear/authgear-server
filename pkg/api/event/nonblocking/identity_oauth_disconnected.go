package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	IdentityOAuthDisconnected event.Type = "identity.oauth.disconnected"
)

type IdentityOAuthDisconnectedEventPayload struct {
	User     model.User     `json:"user"`
	Identity model.Identity `json:"identity"`
	AdminAPI bool           `json:"-"`
}

func (e *IdentityOAuthDisconnectedEventPayload) NonBlockingEventType() event.Type {
	return IdentityOAuthDisconnected
}

func (e *IdentityOAuthDisconnectedEventPayload) UserID() string {
	return e.User.ID
}

func (e *IdentityOAuthDisconnectedEventPayload) IsAdminAPI() bool {
	return e.AdminAPI
}

func (e *IdentityOAuthDisconnectedEventPayload) FillContext(ctx *event.Context) {
}

var _ event.NonBlockingPayload = &IdentityOAuthDisconnectedEventPayload{}

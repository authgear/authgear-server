package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	IdentityOAuthDisconnected event.Type = "identity.oauth.disconnected"
)

type IdentityOAuthDisconnectedEvent struct {
	User     model.User     `json:"user"`
	Identity model.Identity `json:"identity"`
}

func (e *IdentityOAuthDisconnectedEvent) NonBlockingEventType() event.Type {
	return IdentityOAuthDisconnected
}

func (e *IdentityOAuthDisconnectedEvent) UserID() string {
	return e.User.ID
}

var _ event.NonBlockingPayload = &IdentityOAuthDisconnectedEvent{}

package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	IdentityOAuthConnected event.Type = "identity.oauth.connected"
)

type IdentityOAuthConnectedEvent struct {
	User     model.User     `json:"user"`
	Identity model.Identity `json:"identity"`
}

func (e *IdentityOAuthConnectedEvent) NonBlockingEventType() event.Type {
	return IdentityOAuthConnected
}

func (e *IdentityOAuthConnectedEvent) UserID() string {
	return e.User.ID
}

var _ event.NonBlockingPayload = &IdentityOAuthConnectedEvent{}

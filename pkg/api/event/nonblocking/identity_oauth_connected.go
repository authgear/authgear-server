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
	AdminAPI bool           `json:"-"`
}

func (e *IdentityOAuthConnectedEvent) NonBlockingEventType() event.Type {
	return IdentityOAuthConnected
}

func (e *IdentityOAuthConnectedEvent) UserID() string {
	return e.User.ID
}

func (e *IdentityOAuthConnectedEvent) IsAdminAPI() bool {
	return e.AdminAPI
}

var _ event.NonBlockingPayload = &IdentityOAuthConnectedEvent{}

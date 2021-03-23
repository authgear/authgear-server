package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	IdentityCreatedUserAddIdentity     event.Type = "identity.created.user_add_identity"
	IdentityCreatedAdminAPIAddIdentity event.Type = "identity.created.admin_api_add_identity"
)

type IdentityCreatedUserAddIdentityEvent struct {
	User     model.User     `json:"user"`
	Identity model.Identity `json:"identity"`
}

func (e *IdentityCreatedUserAddIdentityEvent) NonBlockingEventType() event.Type {
	return IdentityCreatedUserAddIdentity
}

func (e *IdentityCreatedUserAddIdentityEvent) UserID() string {
	return e.User.ID
}

type IdentityCreatedAdminAPIAddIdentityEvent struct {
	User     model.User     `json:"user"`
	Identity model.Identity `json:"identity"`
}

func (e *IdentityCreatedAdminAPIAddIdentityEvent) NonBlockingEventType() event.Type {
	return IdentityCreatedAdminAPIAddIdentity
}

func (e *IdentityCreatedAdminAPIAddIdentityEvent) UserID() string {
	return e.User.ID
}

var _ event.NonBlockingPayload = &IdentityCreatedUserAddIdentityEvent{}
var _ event.NonBlockingPayload = &IdentityCreatedAdminAPIAddIdentityEvent{}

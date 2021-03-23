package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	IdentityDeletedUserRemoveIdentity     event.Type = "identity.deleted.user_remove_identity"
	IdentityDeletedAdminAPIRemoveIdentity event.Type = "identity.deleted.admin_api_remove_identity"
)

type IdentityDeletedUserRemoveIdentityEvent struct {
	User     model.User     `json:"user"`
	Identity model.Identity `json:"identity"`
}

func (e *IdentityDeletedUserRemoveIdentityEvent) NonBlockingEventType() event.Type {
	return IdentityDeletedUserRemoveIdentity
}

func (e *IdentityDeletedUserRemoveIdentityEvent) UserID() string {
	return e.User.ID
}

type IdentityDeletedAdminAPIRemoveIdentityEvent struct {
	User     model.User     `json:"user"`
	Identity model.Identity `json:"identity"`
}

func (e *IdentityDeletedAdminAPIRemoveIdentityEvent) NonBlockingEventType() event.Type {
	return IdentityDeletedAdminAPIRemoveIdentity
}

func (e *IdentityDeletedAdminAPIRemoveIdentityEvent) UserID() string {
	return e.User.ID
}

var _ event.NonBlockingPayload = &IdentityDeletedUserRemoveIdentityEvent{}
var _ event.NonBlockingPayload = &IdentityDeletedAdminAPIRemoveIdentityEvent{}

package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	IdentityUpdatedUserUpdateIdentity event.Type = "identity.updated.user_update_identity"
)

type IdentityUpdatedUserUpdateIdentityEvent struct {
	User        model.User     `json:"user"`
	NewIdentity model.Identity `json:"new_identity"`
	OldIdentity model.Identity `json:"old_identity"`
}

func (e *IdentityUpdatedUserUpdateIdentityEvent) NonBlockingEventType() event.Type {
	return IdentityUpdatedUserUpdateIdentity
}

func (e *IdentityUpdatedUserUpdateIdentityEvent) UserID() string {
	return e.User.ID
}

var _ event.NonBlockingPayload = &IdentityUpdatedUserUpdateIdentityEvent{}

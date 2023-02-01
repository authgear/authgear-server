package nonblocking

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	IdentityVerifiedFormat string = "identity.%s.verified"
)

type IdentityVerifiedEventPayload struct {
	UserRef   model.UserRef  `json:"-" resolve:"user"`
	UserModel model.User     `json:"user"`
	Identity  model.Identity `json:"identity"`
	ClaimName string         `json:"-"`
	AdminAPI  bool           `json:"-"`
}

func NewIdentityVerifiedEventPayload(
	userRef model.UserRef,
	identity model.Identity,
	claimName string,
	adminAPI bool,
) *IdentityVerifiedEventPayload {
	return &IdentityVerifiedEventPayload{
		UserRef:   userRef,
		Identity:  identity,
		ClaimName: claimName,
		AdminAPI:  adminAPI,
	}
}

func (e *IdentityVerifiedEventPayload) NonBlockingEventType() event.Type {
	return event.Type(fmt.Sprintf(IdentityVerifiedFormat, e.ClaimName))
}

func (e *IdentityVerifiedEventPayload) UserID() string {
	return e.UserRef.ID
}

func (e *IdentityVerifiedEventPayload) GetTriggeredBy() event.TriggeredByType {
	if e.AdminAPI {
		return event.TriggeredByTypeAdminAPI
	}
	return event.TriggeredByTypeUser
}

func (e *IdentityVerifiedEventPayload) FillContext(ctx *event.Context) {
}

func (e *IdentityVerifiedEventPayload) ForHook() bool {
	return true
}

func (e *IdentityVerifiedEventPayload) ForAudit() bool {
	return true
}

func (e *IdentityVerifiedEventPayload) ReindexUserNeeded() bool {
	return false
}

func (e *IdentityVerifiedEventPayload) IsUserDeleted() bool {
	return false
}

var _ event.NonBlockingPayload = &IdentityVerifiedEventPayload{}

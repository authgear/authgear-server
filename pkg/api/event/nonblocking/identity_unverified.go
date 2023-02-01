package nonblocking

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	IdentityUnverifiedFormat string = "identity.%s.unverified"
)

type IdentityUnverifiedEventPayload struct {
	UserRef   model.UserRef  `json:"-" resolve:"user"`
	UserModel model.User     `json:"user"`
	Identity  model.Identity `json:"identity"`
	ClaimName string         `json:"-"`
	AdminAPI  bool           `json:"-"`
}

func NewIdentityUnverifiedEventPayload(
	userRef model.UserRef,
	identity model.Identity,
	claimName string,
	adminAPI bool,
) *IdentityUnverifiedEventPayload {
	return &IdentityUnverifiedEventPayload{
		UserRef:   userRef,
		Identity:  identity,
		ClaimName: claimName,
		AdminAPI:  adminAPI,
	}
}

func (e *IdentityUnverifiedEventPayload) NonBlockingEventType() event.Type {
	return event.Type(fmt.Sprintf(IdentityUnverifiedFormat, e.ClaimName))
}

func (e *IdentityUnverifiedEventPayload) UserID() string {
	return e.UserRef.ID
}

func (e *IdentityUnverifiedEventPayload) GetTriggeredBy() event.TriggeredByType {
	if e.AdminAPI {
		return event.TriggeredByTypeAdminAPI
	}
	return event.TriggeredByTypeUser
}

func (e *IdentityUnverifiedEventPayload) FillContext(ctx *event.Context) {
}

func (e *IdentityUnverifiedEventPayload) ForHook() bool {
	return true
}

func (e *IdentityUnverifiedEventPayload) ForAudit() bool {
	return true
}

func (e *IdentityUnverifiedEventPayload) ReindexUserNeeded() bool {
	return false
}

func (e *IdentityUnverifiedEventPayload) IsUserDeleted() bool {
	return false
}

var _ event.NonBlockingPayload = &IdentityUnverifiedEventPayload{}

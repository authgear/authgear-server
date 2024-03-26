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
	UserRef     model.UserRef  `json:"-" resolve:"user"`
	UserModel   model.User     `json:"user"`
	Identity    model.Identity `json:"identity"`
	LoginIDType string         `json:"-"`
	AdminAPI    bool           `json:"-"`
}

func NewIdentityUnverifiedEventPayload(
	userRef model.UserRef,
	identity model.Identity,
	claimName string,
	adminAPI bool,
) (*IdentityUnverifiedEventPayload, bool) {
	loginIDType, ok := model.GetClaimLoginIDKeyType(model.ClaimName(claimName))
	if !ok {
		return nil, false
	}
	return &IdentityUnverifiedEventPayload{
		UserRef:     userRef,
		Identity:    identity,
		LoginIDType: string(loginIDType),
		AdminAPI:    adminAPI,
	}, true
}

func (e *IdentityUnverifiedEventPayload) NonBlockingEventType() event.Type {
	return event.Type(fmt.Sprintf(IdentityUnverifiedFormat, e.LoginIDType))
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

func (e *IdentityUnverifiedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *IdentityUnverifiedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &IdentityUnverifiedEventPayload{}

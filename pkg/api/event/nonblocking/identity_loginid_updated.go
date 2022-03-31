package nonblocking

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	IdentityLoginIDUpdatedFormat string = "identity.%s.updated"
)

type IdentityLoginIDUpdatedEventPayload struct {
	UserRef     model.UserRef  `json:"-" resolve:"user"`
	UserModel   model.User     `json:"user"`
	NewIdentity model.Identity `json:"new_identity"`
	OldIdentity model.Identity `json:"old_identity"`
	LoginIDType string         `json:"-"`
	AdminAPI    bool           `json:"-"`
}

func NewIdentityLoginIDUpdatedEventPayload(
	userRef model.UserRef,
	newIdentity model.Identity,
	oldIdentity model.Identity,
	loginIDType string,
	adminAPI bool,
) *IdentityLoginIDUpdatedEventPayload {
	if checkIdentityEventSupportLoginIDType(loginIDType) {
		return &IdentityLoginIDUpdatedEventPayload{
			UserRef:     userRef,
			NewIdentity: newIdentity,
			OldIdentity: oldIdentity,
			LoginIDType: loginIDType,
			AdminAPI:    adminAPI,
		}
	}
	return nil
}

func (e *IdentityLoginIDUpdatedEventPayload) NonBlockingEventType() event.Type {
	return event.Type(fmt.Sprintf(IdentityLoginIDUpdatedFormat, e.LoginIDType))
}

func (e *IdentityLoginIDUpdatedEventPayload) UserID() string {
	return e.UserRef.ID
}

func (e *IdentityLoginIDUpdatedEventPayload) GetTriggeredBy() event.TriggeredByType {
	if e.AdminAPI {
		return event.TriggeredByTypeAdminAPI
	}
	return event.TriggeredByTypeUser
}

func (e *IdentityLoginIDUpdatedEventPayload) FillContext(ctx *event.Context) {
}

func (e *IdentityLoginIDUpdatedEventPayload) ForWebHook() bool {
	return true
}

func (e *IdentityLoginIDUpdatedEventPayload) ForAudit() bool {
	return true
}

func (e *IdentityLoginIDUpdatedEventPayload) ReindexUserNeeded() bool {
	return true
}

func (e *IdentityLoginIDUpdatedEventPayload) IsUserDeleted() bool {
	return false
}

var _ event.NonBlockingPayload = &IdentityLoginIDUpdatedEventPayload{}

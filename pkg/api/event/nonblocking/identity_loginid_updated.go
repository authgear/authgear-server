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
) (*IdentityLoginIDUpdatedEventPayload, bool) {
	if !checkIdentityEventSupportLoginIDType(loginIDType) {
		return nil, false
	}
	return &IdentityLoginIDUpdatedEventPayload{
		UserRef:     userRef,
		NewIdentity: newIdentity,
		OldIdentity: oldIdentity,
		LoginIDType: loginIDType,
		AdminAPI:    adminAPI,
	}, true
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

func (e *IdentityLoginIDUpdatedEventPayload) ForHook() bool {
	return true
}

func (e *IdentityLoginIDUpdatedEventPayload) ForAudit() bool {
	return true
}

func (e *IdentityLoginIDUpdatedEventPayload) RequireReindexUserIDs() []string {
	return []string{e.UserID()}
}

func (e *IdentityLoginIDUpdatedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &IdentityLoginIDUpdatedEventPayload{}

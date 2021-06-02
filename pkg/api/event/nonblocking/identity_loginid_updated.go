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
	User        model.User     `json:"user"`
	NewIdentity model.Identity `json:"new_identity"`
	OldIdentity model.Identity `json:"old_identity"`
	LoginIDType string         `json:"-"`
	AdminAPI    bool           `json:"-"`
}

func NewIdentityLoginIDUpdatedEventPayload(
	user model.User,
	newIdentity model.Identity,
	oldIdentity model.Identity,
	loginIDType string,
	adminAPI bool,
) *IdentityLoginIDUpdatedEventPayload {
	if checkIdentityEventSupportLoginIDType(loginIDType) {
		return &IdentityLoginIDUpdatedEventPayload{
			User:        user,
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
	return e.User.ID
}

func (e *IdentityLoginIDUpdatedEventPayload) IsAdminAPI() bool {
	return e.AdminAPI
}

func (e *IdentityLoginIDUpdatedEventPayload) FillContext(ctx *event.Context) {
}

func (e *IdentityLoginIDUpdatedEventPayload) ForWebHook() bool {
	return true
}

func (e *IdentityLoginIDUpdatedEventPayload) ForAudit() bool {
	return true
}

var _ event.NonBlockingPayload = &IdentityLoginIDUpdatedEventPayload{}

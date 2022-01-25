package nonblocking

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	IdentityLoginIDRemovedFormat string = "identity.%s.removed"
)

type IdentityLoginIDRemovedEventPayload struct {
	UserRef     model.UserRef  `json:"-" resolve:"user"`
	UserModel   model.User     `json:"user"`
	Identity    model.Identity `json:"identity"`
	LoginIDType string         `json:"-"`
	AdminAPI    bool           `json:"-"`
}

func NewIdentityLoginIDRemovedEventPayload(
	userRef model.UserRef,
	identity model.Identity,
	loginIDType string,
	adminAPI bool,
) *IdentityLoginIDRemovedEventPayload {
	if checkIdentityEventSupportLoginIDType(loginIDType) {
		return &IdentityLoginIDRemovedEventPayload{
			UserRef:     userRef,
			Identity:    identity,
			LoginIDType: loginIDType,
			AdminAPI:    adminAPI,
		}
	}
	return nil
}

func (e *IdentityLoginIDRemovedEventPayload) NonBlockingEventType() event.Type {
	return event.Type(fmt.Sprintf(IdentityLoginIDRemovedFormat, e.LoginIDType))
}

func (e *IdentityLoginIDRemovedEventPayload) UserID() string {
	return e.UserRef.ID
}

func (e *IdentityLoginIDRemovedEventPayload) GetTriggeredBy() event.TriggeredByType {
	if e.AdminAPI {
		return event.TriggeredByTypeAdminAPI
	}
	return event.TriggeredByTypeUser
}

func (e *IdentityLoginIDRemovedEventPayload) FillContext(ctx *event.Context) {
}

func (e *IdentityLoginIDRemovedEventPayload) ForWebHook() bool {
	return true
}

func (e *IdentityLoginIDRemovedEventPayload) ForAudit() bool {
	return true
}

var _ event.NonBlockingPayload = &IdentityLoginIDRemovedEventPayload{}

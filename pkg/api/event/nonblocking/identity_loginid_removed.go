package nonblocking

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	IdentityLoginIDRemovedFormat string = "identity.%s.removed"
)

type IdentityLoginIDRemovedEvent struct {
	User        model.User     `json:"user"`
	Identity    model.Identity `json:"identity"`
	LoginIDType string         `json:"-"`
	AdminAPI    bool           `json:"-"`
}

func NewIdentityLoginIDRemovedEvent(
	user model.User,
	identity model.Identity,
	loginIDType string,
	adminAPI bool,
) *IdentityLoginIDRemovedEvent {
	if checkIdentityEventSupportLoginIDType(loginIDType) {
		return &IdentityLoginIDRemovedEvent{
			User:        user,
			Identity:    identity,
			LoginIDType: loginIDType,
			AdminAPI:    adminAPI,
		}
	}
	return nil
}

func (e *IdentityLoginIDRemovedEvent) NonBlockingEventType() event.Type {
	return event.Type(fmt.Sprintf(IdentityLoginIDRemovedFormat, e.LoginIDType))
}

func (e *IdentityLoginIDRemovedEvent) UserID() string {
	return e.User.ID
}

func (e *IdentityLoginIDRemovedEvent) IsAdminAPI() bool {
	return e.AdminAPI
}

func (e *IdentityLoginIDRemovedEvent) FillContext(ctx *event.Context) {
}

var _ event.NonBlockingPayload = &IdentityLoginIDRemovedEvent{}

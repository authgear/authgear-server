package nonblocking

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func checkIdentityEventSupportLoginIDType(loginIDType string) bool {
	return loginIDType == string(config.LoginIDKeyTypeEmail) ||
		loginIDType == string(config.LoginIDKeyTypePhone) ||
		loginIDType == string(config.LoginIDKeyTypeUsername)
}

const (
	IdentityLoginIDAddedFormat string = "identity.%s.added"
)

type IdentityLoginIDAddedEventPayload struct {
	User        model.User     `json:"user"`
	Identity    model.Identity `json:"identity"`
	LoginIDType string         `json:"-"`
	AdminAPI    bool           `json:"-"`
}

func NewIdentityLoginIDAddedEventPayload(
	user model.User,
	identity model.Identity,
	loginIDType string,
	adminAPI bool,
) *IdentityLoginIDAddedEventPayload {
	if checkIdentityEventSupportLoginIDType(loginIDType) {
		return &IdentityLoginIDAddedEventPayload{
			User:        user,
			Identity:    identity,
			LoginIDType: loginIDType,
			AdminAPI:    adminAPI,
		}
	}
	return nil
}

func (e *IdentityLoginIDAddedEventPayload) NonBlockingEventType() event.Type {
	return event.Type(fmt.Sprintf(IdentityLoginIDAddedFormat, e.LoginIDType))
}

func (e *IdentityLoginIDAddedEventPayload) UserID() string {
	return e.User.ID
}

func (e *IdentityLoginIDAddedEventPayload) IsAdminAPI() bool {
	return e.AdminAPI
}

func (e *IdentityLoginIDAddedEventPayload) FillContext(ctx *event.Context) {
}

var _ event.NonBlockingPayload = &IdentityLoginIDAddedEventPayload{}

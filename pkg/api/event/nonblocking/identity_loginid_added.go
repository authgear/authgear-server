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
	UserRef     model.UserRef  `json:"-" resolve:"user"`
	UserModel   model.User     `json:"user"`
	Identity    model.Identity `json:"identity"`
	LoginIDType string         `json:"-"`
	AdminAPI    bool           `json:"-"`
}

func NewIdentityLoginIDAddedEventPayload(
	userRef model.UserRef,
	identity model.Identity,
	loginIDType string,
	adminAPI bool,
) *IdentityLoginIDAddedEventPayload {
	if checkIdentityEventSupportLoginIDType(loginIDType) {
		return &IdentityLoginIDAddedEventPayload{
			UserRef:     userRef,
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
	return e.UserRef.ID
}

func (e *IdentityLoginIDAddedEventPayload) GetTriggeredBy() event.TriggeredByType {
	if e.AdminAPI {
		return event.TriggeredByTypeAdminAPI
	}
	return event.TriggeredByTypeUser
}

func (e *IdentityLoginIDAddedEventPayload) FillContext(ctx *event.Context) {
}

func (e *IdentityLoginIDAddedEventPayload) ForWebHook() bool {
	return true
}

func (e *IdentityLoginIDAddedEventPayload) ForAudit() bool {
	return true
}

var _ event.NonBlockingPayload = &IdentityLoginIDAddedEventPayload{}

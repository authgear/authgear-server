package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	IdentityBiometricDisabled event.Type = "identity.biometric.disabled"
)

type IdentityBiometricDisabledEventPayload struct {
	UserRef   model.UserRef  `json:"-" resolve:"user"`
	UserModel model.User     `json:"user"`
	Identity  model.Identity `json:"identity"`
	AdminAPI  bool           `json:"-"`
}

func (e *IdentityBiometricDisabledEventPayload) NonBlockingEventType() event.Type {
	return IdentityBiometricDisabled
}

func (e *IdentityBiometricDisabledEventPayload) UserID() string {
	return e.UserRef.ID
}

func (e *IdentityBiometricDisabledEventPayload) GetTriggeredBy() event.TriggeredByType {
	if e.AdminAPI {
		return event.TriggeredByTypeAdminAPI
	}
	return event.TriggeredByTypeUser
}

func (e *IdentityBiometricDisabledEventPayload) FillContext(ctx *event.Context) {
}

func (e *IdentityBiometricDisabledEventPayload) ForHook() bool {
	return true
}

func (e *IdentityBiometricDisabledEventPayload) ForAudit() bool {
	return true
}

func (e *IdentityBiometricDisabledEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *IdentityBiometricDisabledEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &IdentityBiometricDisabledEventPayload{}

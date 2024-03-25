package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	IdentityBiometricEnabled event.Type = "identity.biometric.enabled"
)

type IdentityBiometricEnabledEventPayload struct {
	UserRef   model.UserRef  `json:"-" resolve:"user"`
	UserModel model.User     `json:"user"`
	Identity  model.Identity `json:"identity"`
	AdminAPI  bool           `json:"-"`
}

func (e *IdentityBiometricEnabledEventPayload) NonBlockingEventType() event.Type {
	return IdentityBiometricEnabled
}

func (e *IdentityBiometricEnabledEventPayload) UserID() string {
	return e.UserRef.ID
}

func (e *IdentityBiometricEnabledEventPayload) GetTriggeredBy() event.TriggeredByType {
	if e.AdminAPI {
		return event.TriggeredByTypeAdminAPI
	}
	return event.TriggeredByTypeUser
}

func (e *IdentityBiometricEnabledEventPayload) FillContext(ctx *event.Context) {
}

func (e *IdentityBiometricEnabledEventPayload) ForHook() bool {
	return true
}

func (e *IdentityBiometricEnabledEventPayload) ForAudit() bool {
	return true
}

func (e *IdentityBiometricEnabledEventPayload) RequireReindexUserIDs() []string {
	// Biometric identity doesn't have any IdentityAwareStandardClaims
	// reindex user is not needed
	return []string{}
}

func (e *IdentityBiometricEnabledEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &IdentityBiometricEnabledEventPayload{}

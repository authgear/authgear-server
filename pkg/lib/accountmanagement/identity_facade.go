package accountmanagement

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type IdentityFacade struct {
	Config       *config.AppConfig
	Store        Store
	Identities   IdentityService
	Events       EventService
	Verification VerificationService
}

func (i *IdentityFacade) MakeLoginIDSpec(loginIDKey string, loginID string) (*identity.Spec, error) {
	// EdgeUseIdentityLoginID
	matchedLoginIDConfig, ok := i.Config.Identity.LoginID.GetKeyConfig(loginIDKey)
	if !ok {
		return nil, api.NewInvariantViolated(
			"InvalidLoginIDKey",
			"invalid login ID key",
			nil,
		)
	}
	typ := matchedLoginIDConfig.Type
	identitySpec := &identity.Spec{
		Type: model.IdentityTypeLoginID,
		LoginID: &identity.LoginIDSpec{
			Key:   loginIDKey,
			Type:  typ,
			Value: loginID,
		},
	}
	return identitySpec, nil
}

func (i *IdentityFacade) CreateIdentity(userID string, identitySpec *identity.Spec, needVerify bool) (*identity.Info, bool, error) {
	// EdgeCreateIdentityEnd
	identityInfo, err := i.Identities.New(userID, identitySpec, identity.NewIdentityOptions{LoginIDEmailByPassBlocklistAllowlist: false})
	if err != nil {
		return nil, false, err
	}
	// EdgeDoCreateIdentity
	createDisabled := identityInfo.CreateDisabled(i.Config.Identity)
	if createDisabled {
		return nil, false, api.ErrIdentityModifyDisabled
	}

	// NodeDoCreateIdentity GetEffects() -> EffectOnCommit()
	if _, err := i.Identities.CheckDuplicated(identityInfo); err != nil {
		return nil, false, err
	}

	if needVerify {
		claims, err := i.Verification.GetIdentityVerificationStatus(identityInfo)
		if err != nil {
			return nil, false, err
		}
		// if not verified, send verification code
		if len(claims) > 0 && !claims[0].Verified {
			return identityInfo, true, nil
		}
	}

	if err = i.Identities.Create(identityInfo); err != nil {
		return nil, false, err
	}

	if err = i.dispatchCreateIdentityEvent(identityInfo); err != nil {
		return nil, false, err
	}

	return identityInfo, false, nil
}

func (i *IdentityFacade) UpdateIdentity(userID string, identityID string, identitySpec *identity.Spec, needVerify bool) (*identity.Info, bool, error) {
	oldInfo, err := i.Identities.Get(identityID)
	fmt.Printf("oldInfo: %v\n", oldInfo)
	if err != nil {
		return nil, false, err
	}

	if oldInfo.UserID != userID {
		return nil, false, ErrAccountManagementIdentityNotOwnedbyToUser
	}

	newInfo, err := i.Identities.UpdateWithSpec(oldInfo, identitySpec, identity.NewIdentityOptions{
		LoginIDEmailByPassBlocklistAllowlist: false,
	})
	if err != nil {
		return nil, false, err
	}

	// EdgeDoUpdateIdentity
	updateDisabled := oldInfo.UpdateDisabled(i.Config.Identity)
	if updateDisabled {
		return nil, false, api.ErrIdentityModifyDisabled
	}

	// NodeDoUpdateIdentity GetEffects() -> EffectRun()
	if _, err := i.Identities.CheckDuplicated(newInfo); err != nil {
		if identity.IsErrDuplicatedIdentity(err) {
			s1 := oldInfo.ToSpec()
			s2 := newInfo.ToSpec()
			return nil, false, identity.NewErrDuplicatedIdentity(&s2, &s1)
		}
		return nil, false, err
	}

	if needVerify {
		claims, err := i.Verification.GetIdentityVerificationStatus(newInfo)
		if err != nil {
			return nil, false, err
		}
		// if not verified, send verification code
		if len(claims) > 0 && !claims[0].Verified {
			return newInfo, true, nil
		}
	}

	// Update identity after verification
	if err := i.Identities.Update(oldInfo, newInfo); err != nil {
		return nil, false, err
	}

	// Dispatch event
	if err = i.dispatchUpdatedIdentityEvent(oldInfo, newInfo); err != nil {
		return nil, false, err
	}

	return newInfo, false, nil
}

func (i *IdentityFacade) RemoveIdentity(userID string, identityID string) (*identity.Info, error) {
	identityInfo, err := i.Identities.Get(identityID)
	fmt.Printf("identityInfo: %v\n", identityInfo)
	if err != nil {
		return nil, err
	}

	if identityInfo.UserID != userID {
		return nil, ErrAccountManagementIdentityNotOwnedbyToUser
	}

	// EdgeDoRemoveIdentity
	deleteDiabled := identityInfo.DeleteDisabled(i.Config.Identity)
	if deleteDiabled {
		return nil, api.ErrIdentityModifyDisabled
	}

	if err := i.Identities.Delete(identityInfo); err != nil {
		return nil, err
	}

	if err := i.dispatchRemoveIdentityEvent(identityInfo); err != nil {
		return nil, err
	}

	return identityInfo, nil
}

func (i *IdentityFacade) dispatchCreateIdentityEvent(identityInfo *identity.Info) (err error) {
	userRef := model.UserRef{
		Meta: model.Meta{
			ID: identityInfo.UserID,
		},
	}

	var e event.Payload
	switch identityInfo.Type {
	case model.IdentityTypeLoginID:
		loginIDType := identityInfo.LoginID.LoginIDType
		if payload, ok := nonblocking.NewIdentityLoginIDAddedEventPayload(
			userRef,
			identityInfo.ToModel(),
			string(loginIDType),
			false,
		); ok {
			e = payload
		}
	case model.IdentityTypeOAuth:
		e = &nonblocking.IdentityOAuthConnectedEventPayload{
			UserRef:  userRef,
			Identity: identityInfo.ToModel(),
			AdminAPI: false,
		}
	case model.IdentityTypeBiometric:
		e = &nonblocking.IdentityBiometricEnabledEventPayload{
			UserRef:  userRef,
			Identity: identityInfo.ToModel(),
			AdminAPI: false,
		}
	}

	if e != nil {
		err = i.Events.DispatchEventOnCommit(e)
		if err != nil {
			return err
		}
	}

	return nil
}

func (i *IdentityFacade) dispatchUpdatedIdentityEvent(identityAfterUpdate *identity.Info, identityBeforeUpdate *identity.Info) (err error) {
	userRef := model.UserRef{
		Meta: model.Meta{
			ID: identityAfterUpdate.UserID,
		},
	}

	var e event.Payload
	switch identityAfterUpdate.Type {
	case model.IdentityTypeLoginID:
		loginIDType := identityAfterUpdate.LoginID.LoginIDType
		if payload, ok := nonblocking.NewIdentityLoginIDUpdatedEventPayload(
			userRef,
			identityAfterUpdate.ToModel(),
			identityBeforeUpdate.ToModel(),
			string(loginIDType),
			false,
		); ok {
			e = payload
		}
	}

	if e != nil {
		err = i.Events.DispatchEventOnCommit(e)
		if err != nil {
			return err
		}
	}

	return nil
}

func (i *IdentityFacade) dispatchRemoveIdentityEvent(identityInfo *identity.Info) (err error) {
	userRef := model.UserRef{
		Meta: model.Meta{
			ID: identityInfo.UserID,
		},
	}

	var e event.Payload
	switch identityInfo.Type {
	case model.IdentityTypeLoginID:
		loginIDType := identityInfo.LoginID.LoginIDType
		if payload, ok := nonblocking.NewIdentityLoginIDRemovedEventPayload(
			userRef,
			identityInfo.ToModel(),
			string(loginIDType),
			false,
		); ok {
			e = payload
		}
	case model.IdentityTypeOAuth:
		e = &nonblocking.IdentityOAuthDisconnectedEventPayload{
			UserRef:  userRef,
			Identity: identityInfo.ToModel(),
			AdminAPI: false,
		}
	case model.IdentityTypeBiometric:
		e = &nonblocking.IdentityBiometricDisabledEventPayload{
			UserRef:  userRef,
			Identity: identityInfo.ToModel(),
			AdminAPI: false,
		}
	}

	if e != nil {
		err = i.Events.DispatchEventOnCommit(e)
		if err != nil {
			return err
		}
	}

	return nil
}

package accountmanagement

import (
	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/session"
)

type AddIdentityUsernameInput struct {
	LoginID    string
	LoginIDKey string
}

type AddIdentityUsernameOutput struct {
	IdentityInfo *identity.Info
}

func (s *Service) AddIdentityUsername(resolvedSession session.ResolvedSession, input *AddIdentityUsernameInput) (*AddIdentityUsernameOutput, error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID
	loginKey := input.LoginIDKey
	loginID := input.LoginID

	spec, err := s.makeLoginIDSpec(loginKey, loginID)
	if err != nil {
		return nil, err
	}

	var info *identity.Info
	err = s.Database.WithTx(func() error {
		info, err = s.prepareNewIdentity(userID, spec)
		if err != nil {
			return err
		}

		err = s.createIdentity(info)
		if err != nil {
			return err
		}

		err = s.dispatchIdentityCreatedEvent(info)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &AddIdentityUsernameOutput{IdentityInfo: info}, nil
}

type UpdateIdentityUsernameInput struct {
	LoginID    string
	LoginIDKey string
	IdentityID string
}

type UpdateIdentityUsernameOutput struct {
	IdentityInfo *identity.Info
}

func (s *Service) UpdateIdentityUsername(resolvedSession session.ResolvedSession, input *UpdateIdentityUsernameInput) (*UpdateIdentityUsernameOutput, error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID
	loginKey := input.LoginIDKey
	loginID := input.LoginID
	identityID := input.IdentityID

	spec, err := s.makeLoginIDSpec(loginKey, loginID)
	if err != nil {
		return nil, err
	}

	var info *identity.Info
	err = s.Database.WithTx(func() error {
		oldInfo, newInfo, err := s.prepareUpdateIdentity(userID, identityID, spec)
		if err != nil {
			return err
		}

		err = s.updateIdentity(oldInfo, newInfo)
		if err != nil {
			return err
		}

		err = s.dispatchIdentityUpdatedEvent(oldInfo, newInfo)
		if err != nil {
			return err
		}

		info = newInfo
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &UpdateIdentityUsernameOutput{IdentityInfo: info}, nil
}

type DeleteIdentityUsernameInput struct {
	IdentityID string
}

type DeleteIdentityUsernameOutput struct {
	IdentityInfo *identity.Info
}

func (s *Service) DeleteIdentityUsername(resolvedSession session.ResolvedSession, input *DeleteIdentityUsernameInput) (*DeleteIdentityUsernameOutput, error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID
	identityID := input.IdentityID

	var info *identity.Info
	err := s.Database.WithTx(func() (err error) {
		info, err = s.prepareDeleteIdentity(userID, identityID)
		if err != nil {
			return err
		}

		err = s.deleteIdentity(info)
		if err != nil {
			return err
		}

		err = s.dispatchIdentityDeletedEvent(info)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &DeleteIdentityUsernameOutput{IdentityInfo: info}, nil
}

func (i *Service) makeLoginIDSpec(loginIDKey string, loginID string) (*identity.Spec, error) {
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

func (s *Service) prepareNewIdentity(userID string, identitySpec *identity.Spec) (*identity.Info, error) {
	info, err := s.Identities.New(userID, identitySpec, identity.NewIdentityOptions{LoginIDEmailByPassBlocklistAllowlist: false})
	if err != nil {
		return nil, err
	}

	createDisabled := info.CreateDisabled(s.Config.Identity)
	if createDisabled {
		return nil, api.ErrIdentityModifyDisabled
	}

	if _, err := s.Identities.CheckDuplicated(info); err != nil {
		if identity.IsErrDuplicatedIdentity(err) {
			return nil, ErrAccountManagementDuplicatedIdentity
		}
		return nil, err
	}

	return info, nil
}

func (s *Service) prepareUpdateIdentity(userID string, identityID string, identitySpec *identity.Spec) (*identity.Info, *identity.Info, error) {
	oldInfo, err := s.Identities.Get(identityID)
	if err != nil {
		return nil, nil, err
	}

	if oldInfo.UserID != userID {
		return nil, nil, ErrAccountManagementIdentityNotOwnedbyToUser
	}

	newInfo, err := s.Identities.UpdateWithSpec(oldInfo, identitySpec, identity.NewIdentityOptions{
		LoginIDEmailByPassBlocklistAllowlist: false,
	})
	if err != nil {
		return nil, nil, err
	}

	updateDisabled := oldInfo.UpdateDisabled(s.Config.Identity)
	if updateDisabled {
		return nil, nil, api.ErrIdentityModifyDisabled
	}

	if _, err := s.Identities.CheckDuplicated(newInfo); err != nil {
		if identity.IsErrDuplicatedIdentity(err) {
			return nil, nil, ErrAccountManagementDuplicatedIdentity
		}
		return nil, nil, err
	}

	return oldInfo, newInfo, nil
}

func (s *Service) prepareDeleteIdentity(userID string, identityID string) (*identity.Info, error) {
	info, err := s.Identities.Get(identityID)
	if err != nil {
		return nil, err
	}

	if info.UserID != userID {
		return nil, ErrAccountManagementIdentityNotOwnedbyToUser
	}

	deleteDiabled := info.DeleteDisabled(s.Config.Identity)
	if deleteDiabled {
		return nil, api.ErrIdentityModifyDisabled
	}

	return info, nil
}

func (s *Service) checkIdentityVerified(info *identity.Info) (bool, error) {
	claims, err := s.Verification.GetIdentityVerificationStatus(info)
	if err != nil {
		return false, err
	}
	if len(claims) == 0 {
		return false, nil
	}
	claim := claims[0]
	return claim.Verified, nil
}

func (s *Service) createIdentity(info *identity.Info) error {
	return s.Identities.Create(info)
}

func (s *Service) updateIdentity(oldInfo *identity.Info, newInfo *identity.Info) error {
	return s.Identities.Update(oldInfo, newInfo)
}

func (s *Service) deleteIdentity(info *identity.Info) error {
	return s.Identities.Delete(info)
}

func (s *Service) dispatchIdentityCreatedEvent(info *identity.Info) (err error) {
	userRef := model.UserRef{
		Meta: model.Meta{
			ID: info.UserID,
		},
	}

	var e event.Payload
	switch info.Type {
	case model.IdentityTypeLoginID:
		loginIDType := info.LoginID.LoginIDType
		if payload, ok := nonblocking.NewIdentityLoginIDAddedEventPayload(
			userRef,
			info.ToModel(),
			string(loginIDType),
			false,
		); ok {
			e = payload
		}
	case model.IdentityTypeOAuth:
		e = &nonblocking.IdentityOAuthConnectedEventPayload{
			UserRef:  userRef,
			Identity: info.ToModel(),
			AdminAPI: false,
		}
	case model.IdentityTypeBiometric:
		e = &nonblocking.IdentityBiometricEnabledEventPayload{
			UserRef:  userRef,
			Identity: info.ToModel(),
			AdminAPI: false,
		}
	}

	if e != nil {
		err = s.Events.DispatchEventOnCommit(e)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) dispatchIdentityUpdatedEvent(oldInfo *identity.Info, newInfo *identity.Info) (err error) {
	userRef := model.UserRef{
		Meta: model.Meta{
			ID: newInfo.UserID,
		},
	}

	var e event.Payload
	switch newInfo.Type {
	case model.IdentityTypeLoginID:
		loginIDType := newInfo.LoginID.LoginIDType
		if payload, ok := nonblocking.NewIdentityLoginIDUpdatedEventPayload(
			userRef,
			newInfo.ToModel(),
			oldInfo.ToModel(),
			string(loginIDType),
			false,
		); ok {
			e = payload
		}
	}

	if e != nil {
		err = s.Events.DispatchEventOnCommit(e)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) dispatchIdentityDeletedEvent(info *identity.Info) (err error) {
	userRef := model.UserRef{
		Meta: model.Meta{
			ID: info.UserID,
		},
	}

	var e event.Payload
	switch info.Type {
	case model.IdentityTypeLoginID:
		loginIDType := info.LoginID.LoginIDType
		if payload, ok := nonblocking.NewIdentityLoginIDRemovedEventPayload(
			userRef,
			info.ToModel(),
			string(loginIDType),
			false,
		); ok {
			e = payload
		}
	case model.IdentityTypeOAuth:
		e = &nonblocking.IdentityOAuthDisconnectedEventPayload{
			UserRef:  userRef,
			Identity: info.ToModel(),
			AdminAPI: false,
		}
	case model.IdentityTypeBiometric:
		e = &nonblocking.IdentityBiometricDisabledEventPayload{
			UserRef:  userRef,
			Identity: info.ToModel(),
			AdminAPI: false,
		}
	}

	if e != nil {
		err = s.Events.DispatchEventOnCommit(e)
		if err != nil {
			return err
		}
	}

	return nil
}

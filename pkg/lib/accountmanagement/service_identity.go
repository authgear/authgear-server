package accountmanagement

import (
	"encoding/json"
	"fmt"

	"github.com/go-webauthn/webauthn/protocol"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/uuid"
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

type StartAddIdentityEmailInput struct {
	LoginID    string
	LoginIDKey string
}

type StartAddIdentityEmailOutput struct {
	IdentityInfo     *identity.Info
	NeedVerification bool
	Token            string
}

func (s *Service) StartAddIdentityEmail(resolvedSession session.ResolvedSession, input *StartAddIdentityEmailInput) (*StartAddIdentityEmailOutput, error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID
	loginKey := input.LoginIDKey
	loginID := input.LoginID

	spec, err := s.makeLoginIDSpec(loginKey, loginID)
	if err != nil {
		return nil, err
	}

	var info *identity.Info
	var token string
	var needVerification bool
	err = s.Database.WithTx(func() error {
		info, err = s.prepareNewIdentity(userID, spec)
		if err != nil {
			return err
		}

		verified, err := s.CheckIdentityVerified(info)
		if err != nil {
			return err
		}
		needVerification = !verified
		if needVerification {
			channel, target := info.LoginID.ToChannelTarget()
			err = s.sendOTPCode(userID, channel, target, false)
			if err != nil {
				return err
			}
			token, err = s.Store.GenerateToken(GenerateTokenOptions{
				UserID: userID,
				Email:  info.LoginID.LoginID,
			})
			if err != nil {
				return err
			}
			return nil
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

	return &StartAddIdentityEmailOutput{
		IdentityInfo:     info,
		NeedVerification: needVerification,
		Token:            token,
	}, nil
}

type ResumeAddIdentityEmailInput struct {
	LoginIDKey string
	Token      string
	Code       string
}

type ResumeAddIdentityEmailOutput struct {
	IdentityInfo *identity.Info
}

func (s *Service) ResumeAddIdentityEmail(resolvedSession session.ResolvedSession, input *ResumeAddIdentityEmailInput) (output *ResumeAddIdentityEmailOutput, err error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID
	token, err := s.Store.GetToken(input.Token)
	defer func() {
		if err == nil {
			_, err = s.Store.ConsumeToken(input.Token)
		}
	}()

	if err != nil {
		return
	}
	err = token.CheckUser(userID)
	if err != nil {
		return
	}

	err = s.verifyOTP(userID, model.AuthenticatorOOBChannelEmail, token.Identity.Email, input.Code)
	if err != nil {
		return
	}

	spec, err := s.makeLoginIDSpec(input.LoginIDKey, token.Identity.Email)
	if err != nil {
		return
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

		claimName, ok := model.GetLoginIDKeyTypeClaim(info.LoginID.LoginIDType)
		if !ok {
			panic(fmt.Errorf("accountmanagement: unexpected login ID key"))
		}
		err = s.markClaimVerified(userID, claimName, info.LoginID.LoginID)
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

	output = &ResumeAddIdentityEmailOutput{
		IdentityInfo: info,
	}
	return
}

type StartUpdateIdentityEmailInput struct {
	IdentityID string
	LoginID    string
	LoginIDKey string
}

type StartUpdateIdentityEmailOutput struct {
	OldInfo          *identity.Info
	NewInfo          *identity.Info
	NeedVerification bool
	Token            string
}

func (s *Service) StartUpdateIdentityEmail(resolvedSession session.ResolvedSession, input *StartUpdateIdentityEmailInput) (*StartUpdateIdentityEmailOutput, error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID
	loginKey := input.LoginIDKey
	loginID := input.LoginID

	spec, err := s.makeLoginIDSpec(loginKey, loginID)
	if err != nil {
		return nil, err
	}

	var oldInfo *identity.Info
	var newInfo *identity.Info
	var token string
	var needVerification bool
	err = s.Database.WithTx(func() error {
		oldInfo, newInfo, err := s.prepareUpdateIdentity(userID, input.IdentityID, spec)
		if err != nil {
			return err
		}

		verified, err := s.CheckIdentityVerified(newInfo)
		if err != nil {
			return err
		}
		needVerification = !verified

		if needVerification {
			channel, target := newInfo.LoginID.ToChannelTarget()
			err = s.sendOTPCode(userID, channel, target, false)
			if err != nil {
				return err
			}
			token, err = s.Store.GenerateToken(GenerateTokenOptions{
				UserID:     userID,
				Email:      newInfo.LoginID.LoginID,
				IdentityID: newInfo.ID,
			})
			if err != nil {
				return err
			}
			return nil
		}

		err = s.updateIdentity(oldInfo, newInfo)
		if err != nil {
			return err
		}

		err = s.dispatchIdentityCreatedEvent(newInfo)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &StartUpdateIdentityEmailOutput{
		OldInfo:          oldInfo,
		NewInfo:          newInfo,
		NeedVerification: needVerification,
		Token:            token,
	}, nil
}

type ResumeUpdateIdentityEmailInput struct {
	LoginIDKey string
	Token      string
	Code       string
}

type ResumeUpdateIdentityEmailOutput struct {
	OldInfo *identity.Info
	NewInfo *identity.Info
}

func (s *Service) ResumeUpdateIdentityEmail(resolvedSession session.ResolvedSession, input *ResumeAddIdentityEmailInput) (output *ResumeUpdateIdentityEmailOutput, err error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID
	token, err := s.Store.GetToken(input.Token)
	defer func() {
		if err == nil {
			_, err = s.Store.ConsumeToken(input.Token)
		}
	}()

	if err != nil {
		return
	}
	err = token.CheckUser(userID)
	if err != nil {
		return
	}

	err = s.verifyOTP(userID, model.AuthenticatorOOBChannelEmail, token.Identity.Email, input.Code)
	if err != nil {
		return
	}

	spec, err := s.makeLoginIDSpec(input.LoginIDKey, token.Identity.Email)
	if err != nil {
		return
	}

	var oldInfo *identity.Info
	var newInfo *identity.Info
	err = s.Database.WithTx(func() error {
		oldInfo, newInfo, err = s.prepareUpdateIdentity(userID, token.Identity.IdentityID, spec)
		if err != nil {
			return err
		}

		err = s.updateIdentity(oldInfo, newInfo)
		if err != nil {
			return err
		}

		claimName, ok := model.GetLoginIDKeyTypeClaim(newInfo.LoginID.LoginIDType)
		if !ok {
			panic(fmt.Errorf("accountmanagement: unexpected login ID key"))
		}
		err = s.markClaimVerified(userID, claimName, newInfo.LoginID.LoginID)
		if err != nil {
			return err
		}

		err = s.dispatchIdentityUpdatedEvent(oldInfo, newInfo)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	output = &ResumeUpdateIdentityEmailOutput{
		OldInfo: oldInfo,
		NewInfo: newInfo,
	}
	return
}

type ResumeAddOrUpdateIdentityEmailInput struct {
	LoginIDKey string
	Code       string
}

type ResumeAddOrUpdateIdentityEmailOutput struct {
	OldInfo *identity.Info
	NewInfo *identity.Info
}

func (s *Service) ResumeAddOrUpdateIdentityEmail(resolvedSession session.ResolvedSession, tokenString string, input *ResumeAddOrUpdateIdentityEmailInput) (*ResumeAddOrUpdateIdentityEmailOutput, error) {
	token, err := s.Store.GetToken(tokenString)
	if err != nil {
		return nil, err
	}

	if token.Identity.IdentityID == "" {
		output, err := s.ResumeAddIdentityEmail(resolvedSession, tokenString, &ResumeAddIdentityEmailInput{
			LoginIDKey: input.LoginIDKey,
			Code:       input.Code,
		})
		if err != nil {
			return nil, err
		}
		return &ResumeAddOrUpdateIdentityEmailOutput{
			NewInfo: output.IdentityInfo,
		}, nil
	}

	output, err := s.ResumeUpdateIdentityEmail(resolvedSession, tokenString, &ResumeUpdateIdentityEmailInput{
		LoginIDKey: input.LoginIDKey,
		Code:       input.Code,
	})
	if err != nil {
		return nil, err
	}
	return &ResumeAddOrUpdateIdentityEmailOutput{
		OldInfo: output.OldInfo,
		NewInfo: output.NewInfo,
	}, nil
}

type DeleteIdentityEmailInput struct {
	IdentityID string
}

type DeleteIdentityEmailOutput struct {
	IdentityInfo *identity.Info
}

func (s *Service) DeleteIdentityEmail(resolvedSession session.ResolvedSession, input *DeleteIdentityEmailInput) (*DeleteIdentityEmailOutput, error) {
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

	return &DeleteIdentityEmailOutput{IdentityInfo: info}, nil
}

type StartAddIdentityPhoneInput struct {
	LoginID    string
	LoginIDKey string
}

type StartAddIdentityPhoneOutput struct {
	IdentityInfo     *identity.Info
	NeedVerification bool
	Token            string
}

func (s *Service) StartAddIdentityPhone(resolvedSession session.ResolvedSession, input *StartAddIdentityPhoneInput) (*StartAddIdentityPhoneOutput, error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID
	loginKey := input.LoginIDKey
	loginID := input.LoginID

	spec, err := s.makeLoginIDSpec(loginKey, loginID)
	if err != nil {
		return nil, err
	}

	var info *identity.Info
	var token string
	var needVerification bool
	err = s.Database.WithTx(func() error {
		info, err = s.prepareNewIdentity(userID, spec)
		if err != nil {
			return err
		}

		verified, err := s.CheckIdentityVerified(info)
		if err != nil {
			return err
		}
		needVerification = !verified

		if needVerification {
			channel, target := info.LoginID.ToChannelTarget()
			err = s.sendOTPCode(userID, channel, target, false)
			if err != nil {
				return err
			}
			token, err = s.Store.GenerateToken(GenerateTokenOptions{
				UserID:      userID,
				PhoneNumber: info.LoginID.LoginID,
			})
			if err != nil {
				return err
			}
			return nil
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

	return &StartAddIdentityPhoneOutput{
		IdentityInfo:     info,
		NeedVerification: needVerification,
		Token:            token,
	}, nil
}

type ResumeAddIdentityPhoneInput struct {
	LoginIDKey string
	Token      string
	Code       string
}

type ResumeAddIdentityPhoneOutput struct {
	IdentityInfo *identity.Info
}

func (s *Service) ResumeAddIdentityPhone(resolvedSession session.ResolvedSession, input *ResumeAddIdentityPhoneInput) (output *ResumeAddIdentityPhoneOutput, err error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID
	token, err := s.Store.GetToken(input.Token)
	defer func() {
		if err == nil {
			_, err = s.Store.ConsumeToken(input.Token)
		}
	}()

	if err != nil {
		return
	}
	err = token.CheckUser(userID)
	if err != nil {
		return
	}

	err = s.verifyOTP(userID, model.AuthenticatorOOBChannelSMS, token.Identity.PhoneNumber, input.Code)
	if err != nil {
		return
	}

	spec, err := s.makeLoginIDSpec(input.LoginIDKey, token.Identity.PhoneNumber)
	if err != nil {
		return
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

		claimName, ok := model.GetLoginIDKeyTypeClaim(info.LoginID.LoginIDType)
		if !ok {
			panic(fmt.Errorf("accountmanagement: unexpected login ID key"))
		}
		err = s.markClaimVerified(userID, claimName, info.LoginID.LoginID)
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

	output = &ResumeAddIdentityPhoneOutput{
		IdentityInfo: info,
	}
	return
}

type StartUpdateIdentityPhoneInput struct {
	IdentityID string
	LoginID    string
	LoginIDKey string
}

type StartUpdateIdentityPhoneOutput struct {
	IdentityInfo     *identity.Info
	NeedVerification bool
	Token            string
}

func (s *Service) StartUpdateIdentityPhone(resolvedSession session.ResolvedSession, input *StartUpdateIdentityPhoneInput) (*StartUpdateIdentityPhoneOutput, error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID
	loginKey := input.LoginIDKey
	loginID := input.LoginID

	spec, err := s.makeLoginIDSpec(loginKey, loginID)
	if err != nil {
		return nil, err
	}

	var info *identity.Info
	var token string
	var needVerification bool
	err = s.Database.WithTx(func() error {
		oldInfo, newInfo, err := s.prepareUpdateIdentity(userID, input.IdentityID, spec)
		if err != nil {
			return err
		}
		info = newInfo

		verified, err := s.CheckIdentityVerified(newInfo)
		if err != nil {
			return err
		}
		needVerification = !verified
		if needVerification {
			channel, target := info.LoginID.ToChannelTarget()
			err = s.sendOTPCode(userID, channel, target, false)
			if err != nil {
				return err
			}
			token, err = s.Store.GenerateToken(GenerateTokenOptions{
				UserID:      userID,
				PhoneNumber: info.LoginID.LoginID,
				IdentityID:  info.ID,
			})
			if err != nil {
				return err
			}
			return nil
		}

		err = s.updateIdentity(oldInfo, newInfo)
		if err != nil {
			return err
		}

		err = s.dispatchIdentityCreatedEvent(newInfo)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &StartUpdateIdentityPhoneOutput{
		IdentityInfo:     info,
		NeedVerification: needVerification,
		Token:            token,
	}, nil
}

type ResumeUpdateIdentityPhoneInput struct {
	LoginIDKey string
	Token      string
	Code       string
}

type ResumeUpdateIdentityPhoneOutput struct {
	IdentityInfo *identity.Info
}

func (s *Service) ResumeUpdateIdentityPhone(resolvedSession session.ResolvedSession, input *ResumeAddIdentityEmailInput) (output *ResumeUpdateIdentityPhoneOutput, err error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID
	token, err := s.Store.GetToken(input.Token)
	defer func() {
		if err == nil {
			_, err = s.Store.ConsumeToken(input.Token)
		}
	}()

	if err != nil {
		return
	}
	err = token.CheckUser(userID)
	if err != nil {
		return
	}

	err = s.verifyOTP(userID, model.AuthenticatorOOBChannelSMS, token.Identity.PhoneNumber, input.Code)
	if err != nil {
		return
	}

	spec, err := s.makeLoginIDSpec(input.LoginIDKey, token.Identity.PhoneNumber)
	if err != nil {
		return
	}

	var info *identity.Info
	err = s.Database.WithTx(func() error {
		oldInfo, newInfo, err := s.prepareUpdateIdentity(userID, token.Identity.IdentityID, spec)
		if err != nil {
			return err
		}
		info = newInfo

		err = s.updateIdentity(oldInfo, newInfo)
		if err != nil {
			return err
		}

		claimName, ok := model.GetLoginIDKeyTypeClaim(info.LoginID.LoginIDType)
		if !ok {
			panic(fmt.Errorf("accountmanagement: unexpected login ID key"))
		}
		err = s.markClaimVerified(userID, claimName, info.LoginID.LoginID)
		if err != nil {
			return err
		}

		err = s.dispatchIdentityUpdatedEvent(oldInfo, newInfo)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	output = &ResumeUpdateIdentityPhoneOutput{
		IdentityInfo: info,
	}
	return
}

type DeleteIdentityPhoneInput struct {
	IdentityID string
}

type DeleteIdentityPhoneOutput struct {
	IdentityInfo *identity.Info
}

func (s *Service) DeleteIdentityPhone(resolvedSession session.ResolvedSession, input *DeleteIdentityPhoneInput) (*DeleteIdentityPhoneOutput, error) {
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

	return &DeleteIdentityPhoneOutput{IdentityInfo: info}, nil
}

type AddPasskeyInput struct {
	CreationResponse *protocol.CredentialCreationResponse
}

type AddPasskeyOutput struct {
	IdentityInfo *identity.Info
}

func (s *Service) AddPasskey(resolvedSession session.ResolvedSession, input *AddPasskeyInput) (*AddPasskeyOutput, error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID
	creationResponse := input.CreationResponse
	creationResponseBytes, err := json.Marshal(creationResponse)
	if err != nil {
		return nil, err
	}

	authenticatorSpec := &authenticator.Spec{
		UserID: userID,
		Kind:   authenticator.KindPrimary,
		Type:   model.AuthenticatorTypePasskey,
		Passkey: &authenticator.PasskeySpec{
			AttestationResponse: creationResponseBytes,
		},
	}

	authenticatorID := uuid.New()
	authenticatorInfo, err := s.Authenticators.NewWithAuthenticatorID(authenticatorID, authenticatorSpec)
	if err != nil {
		return nil, err
	}

	identitySpec := &identity.Spec{
		Type: model.IdentityTypePasskey,
		Passkey: &identity.PasskeySpec{
			AttestationResponse: creationResponseBytes,
		},
	}

	var identityInfo *identity.Info
	err = s.Database.WithTx(func() error {
		identityInfo, err = s.prepareNewIdentity(userID, identitySpec)
		if err != nil {
			return err
		}

		err = s.createIdentity(identityInfo)
		if err != nil {
			return err
		}

		err = s.dispatchIdentityCreatedEvent(identityInfo)
		if err != nil {
			return err
		}

		err = s.Authenticators.Create(authenticatorInfo, false)
		if err != nil {
			return err
		}
		err = s.PasskeyService.ConsumeAttestationResponse(creationResponseBytes)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &AddPasskeyOutput{IdentityInfo: identityInfo}, nil
}

type DeletePasskeyInput struct {
	IdentityID string
}

type DeletePasskeyOutput struct {
	IdentityInfo *identity.Info
}

func (s *Service) DeletePasskey(resolvedSession session.ResolvedSession, input *DeletePasskeyInput) (*DeletePasskeyOutput, error) {
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

	return &DeletePasskeyOutput{IdentityInfo: info}, nil
}

type DeleteIdentityBiometricInput struct {
	IdentityID string
}

type DeleteIdentityBiometricOuput struct {
	IdentityInfo *identity.Info
}

func (s *Service) DeleteIdentityBiometric(resolvedSession session.ResolvedSession, input *DeleteIdentityBiometricInput) (*DeleteIdentityBiometricOuput, error) {
	identityID := input.IdentityID
	userID := resolvedSession.GetAuthenticationInfo().UserID

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

	return &DeleteIdentityBiometricOuput{IdentityInfo: info}, nil
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

func (s *Service) CheckIdentityVerified(info *identity.Info) (bool, error) {
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

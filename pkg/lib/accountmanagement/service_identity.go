package accountmanagement

import (
	"context"
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
	"github.com/authgear/authgear-server/pkg/util/stringutil"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

type AddIdentityUsernameInput struct {
	LoginID    string
	LoginIDKey string
}

type AddIdentityUsernameOutput struct {
	IdentityInfo *identity.Info
}

func (s *Service) AddIdentityUsername(ctx context.Context, resolvedSession session.ResolvedSession, input *AddIdentityUsernameInput) (*AddIdentityUsernameOutput, error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID
	loginKey := input.LoginIDKey
	loginID := input.LoginID

	spec, err := s.makeLoginIDSpec(loginKey, loginID)
	if err != nil {
		return nil, err
	}

	var info *identity.Info
	err = s.Database.WithTx(ctx, func(ctx context.Context) error {
		info, err = s.prepareNewIdentity(ctx, userID, spec)
		if err != nil {
			return err
		}

		err = s.createIdentity(ctx, info)
		if err != nil {
			return err
		}

		err = s.dispatchIdentityCreatedEvent(ctx, info)
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

func (s *Service) UpdateIdentityUsername(ctx context.Context, resolvedSession session.ResolvedSession, input *UpdateIdentityUsernameInput) (*UpdateIdentityUsernameOutput, error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID
	loginKey := input.LoginIDKey
	loginID := input.LoginID
	identityID := input.IdentityID

	spec, err := s.makeLoginIDSpec(loginKey, loginID)
	if err != nil {
		return nil, err
	}

	var info *identity.Info
	err = s.Database.WithTx(ctx, func(ctx context.Context) error {
		oldInfo, newInfo, err := s.prepareUpdateIdentity(ctx, userID, identityID, spec)
		if err != nil {
			return err
		}

		err = s.updateIdentity(ctx, oldInfo, newInfo)
		if err != nil {
			return err
		}

		err = s.dispatchIdentityUpdatedEvent(ctx, oldInfo, newInfo)
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

func (s *Service) DeleteIdentityUsername(ctx context.Context, resolvedSession session.ResolvedSession, input *DeleteIdentityUsernameInput) (*DeleteIdentityUsernameOutput, error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID
	identityID := input.IdentityID

	var info *identity.Info
	err := s.Database.WithTx(ctx, func(ctx context.Context) (err error) {
		info, err = s.prepareDeleteIdentity(ctx, userID, identityID)
		if err != nil {
			return err
		}

		err = s.deleteIdentity(ctx, info)
		if err != nil {
			return err
		}

		err = s.dispatchIdentityDeletedEvent(ctx, info)
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

func (s *Service) StartAddIdentityEmail(ctx context.Context, resolvedSession session.ResolvedSession, input *StartAddIdentityEmailInput) (*StartAddIdentityEmailOutput, error) {
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
	err = s.Database.WithTx(ctx, func(ctx context.Context) error {
		info, err = s.prepareNewIdentity(ctx, userID, spec)
		if err != nil {
			return err
		}

		verified, err := s.CheckIdentityVerified(ctx, info)
		if err != nil {
			return err
		}
		needVerification = !verified && *s.Config.Verification.Claims.Email.Enabled && *s.Config.Verification.Claims.Email.Required
		if needVerification {
			target := info.LoginID.LoginID
			channel := model.AuthenticatorOOBChannelEmail
			err = s.sendOTPCode(ctx, userID, channel, target, false)
			if err != nil {
				return err
			}
			token, err = s.Store.GenerateToken(ctx, GenerateTokenOptions{
				UserID:        userID,
				IdentityEmail: info.LoginID.LoginID,
			})
			if err != nil {
				return err
			}
			return nil
		}

		err = s.createIdentity(ctx, info)
		if err != nil {
			return err
		}

		err = s.dispatchIdentityCreatedEvent(ctx, info)
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
	Code       string
}

type ResumeAddIdentityEmailOutput struct {
	IdentityInfo *identity.Info
}

func (s *Service) ResumeAddIdentityEmail(ctx context.Context, resolvedSession session.ResolvedSession, tokenString string, input *ResumeAddIdentityEmailInput) (output *ResumeAddIdentityEmailOutput, err error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID
	token, err := s.Store.GetToken(ctx, tokenString)
	defer func() {
		if err == nil {
			_, err = s.Store.ConsumeToken(ctx, tokenString)
		}
	}()

	if err != nil {
		return
	}
	err = token.CheckUser(userID)
	if err != nil {
		return
	}

	err = s.VerifyOTP(ctx, userID, model.AuthenticatorOOBChannelEmail, token.Identity.Email, input.Code, false)
	if err != nil {
		return
	}

	spec, err := s.makeLoginIDSpec(input.LoginIDKey, token.Identity.Email)
	if err != nil {
		return
	}

	var info *identity.Info
	err = s.Database.WithTx(ctx, func(ctx context.Context) error {
		info, err = s.prepareNewIdentity(ctx, userID, spec)
		if err != nil {
			return err
		}

		err = s.createIdentity(ctx, info)
		if err != nil {
			return err
		}

		claimName, ok := model.GetLoginIDKeyTypeClaim(info.LoginID.LoginIDType)
		if !ok {
			panic(fmt.Errorf("accountmanagement: unexpected login ID key"))
		}
		err = s.markClaimVerified(ctx, userID, claimName, info.LoginID.LoginID)
		if err != nil {
			return err
		}

		err = s.dispatchIdentityCreatedEvent(ctx, info)
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

func (s *Service) StartUpdateIdentityEmail(ctx context.Context, resolvedSession session.ResolvedSession, input *StartUpdateIdentityEmailInput) (*StartUpdateIdentityEmailOutput, error) {
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
	err = s.Database.WithTx(ctx, func(ctx context.Context) error {
		oldInfo, newInfo, err := s.prepareUpdateIdentity(ctx, userID, input.IdentityID, spec)
		if err != nil {
			return err
		}

		verified, err := s.CheckIdentityVerified(ctx, newInfo)
		if err != nil {
			return err
		}
		needVerification = !verified && *s.Config.Verification.Claims.Email.Enabled && *s.Config.Verification.Claims.Email.Required

		if needVerification {
			target := newInfo.LoginID.LoginID
			channel := model.AuthenticatorOOBChannelEmail
			err = s.sendOTPCode(ctx, userID, channel, target, false)
			if err != nil {
				return err
			}
			token, err = s.Store.GenerateToken(ctx, GenerateTokenOptions{
				UserID:        userID,
				IdentityEmail: newInfo.LoginID.LoginID,
				IdentityID:    newInfo.ID,
			})
			if err != nil {
				return err
			}
			return nil
		}

		err = s.updateIdentity(ctx, oldInfo, newInfo)
		if err != nil {
			return err
		}

		err = s.dispatchIdentityCreatedEvent(ctx, newInfo)
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
	Code       string
}

type ResumeUpdateIdentityEmailOutput struct {
	OldInfo *identity.Info
	NewInfo *identity.Info
}

func (s *Service) ResumeUpdateIdentityEmail(ctx context.Context, resolvedSession session.ResolvedSession, tokenString string, input *ResumeUpdateIdentityEmailInput) (output *ResumeUpdateIdentityEmailOutput, err error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID
	token, err := s.Store.GetToken(ctx, tokenString)
	defer func() {
		if err == nil {
			_, err = s.Store.ConsumeToken(ctx, tokenString)
		}
	}()

	if err != nil {
		return
	}
	err = token.CheckUser(userID)
	if err != nil {
		return
	}

	err = s.VerifyOTP(ctx, userID, model.AuthenticatorOOBChannelEmail, token.Identity.Email, input.Code, false)
	if err != nil {
		return
	}

	spec, err := s.makeLoginIDSpec(input.LoginIDKey, token.Identity.Email)
	if err != nil {
		return
	}

	var oldInfo *identity.Info
	var newInfo *identity.Info
	err = s.Database.WithTx(ctx, func(ctx context.Context) error {
		oldInfo, newInfo, err = s.prepareUpdateIdentity(ctx, userID, token.Identity.IdentityID, spec)
		if err != nil {
			return err
		}

		err = s.updateIdentity(ctx, oldInfo, newInfo)
		if err != nil {
			return err
		}

		claimName, ok := model.GetLoginIDKeyTypeClaim(newInfo.LoginID.LoginIDType)
		if !ok {
			panic(fmt.Errorf("accountmanagement: unexpected login ID key"))
		}
		err = s.markClaimVerified(ctx, userID, claimName, newInfo.LoginID.LoginID)
		if err != nil {
			return err
		}

		err = s.dispatchIdentityUpdatedEvent(ctx, oldInfo, newInfo)
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

func (s *Service) ResumeAddOrUpdateIdentityEmail(ctx context.Context, resolvedSession session.ResolvedSession, tokenString string, input *ResumeAddOrUpdateIdentityEmailInput) (*ResumeAddOrUpdateIdentityEmailOutput, error) {
	token, err := s.Store.GetToken(ctx, tokenString)
	if err != nil {
		return nil, err
	}

	if token.Identity.IdentityID == "" {
		output, err := s.ResumeAddIdentityEmail(ctx, resolvedSession, tokenString, &ResumeAddIdentityEmailInput{
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

	output, err := s.ResumeUpdateIdentityEmail(ctx, resolvedSession, tokenString, &ResumeUpdateIdentityEmailInput{
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

func (s *Service) DeleteIdentityEmail(ctx context.Context, resolvedSession session.ResolvedSession, input *DeleteIdentityEmailInput) (*DeleteIdentityEmailOutput, error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID
	identityID := input.IdentityID

	var info *identity.Info
	err := s.Database.WithTx(ctx, func(ctx context.Context) (err error) {
		info, err = s.prepareDeleteIdentity(ctx, userID, identityID)
		if err != nil {
			return err
		}

		err = s.deleteIdentity(ctx, info)
		if err != nil {
			return err
		}

		err = s.dispatchIdentityDeletedEvent(ctx, info)
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
	Channel    model.AuthenticatorOOBChannel
	LoginID    string
	LoginIDKey string
}

type StartAddIdentityPhoneOutput struct {
	IdentityInfo     *identity.Info
	NeedVerification bool
	Token            string
}

func (s *Service) StartAddIdentityPhone(ctx context.Context, resolvedSession session.ResolvedSession, input *StartAddIdentityPhoneInput) (*StartAddIdentityPhoneOutput, error) {
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
	err = s.Database.WithTx(ctx, func(ctx context.Context) error {
		info, err = s.prepareNewIdentity(ctx, userID, spec)
		if err != nil {
			return err
		}

		verified, err := s.CheckIdentityVerified(ctx, info)
		if err != nil {
			return err
		}
		needVerification = !verified && *s.Config.Verification.Claims.PhoneNumber.Enabled && *s.Config.Verification.Claims.PhoneNumber.Required

		if needVerification {
			target := info.LoginID.LoginID
			err = s.sendOTPCode(ctx, userID, input.Channel, target, false)
			if err != nil {
				return err
			}
			token, err = s.Store.GenerateToken(ctx, GenerateTokenOptions{
				UserID:              userID,
				IdentityChannel:     input.Channel,
				IdentityPhoneNumber: info.LoginID.LoginID,
			})
			if err != nil {
				return err
			}
			return nil
		}

		err = s.createIdentity(ctx, info)
		if err != nil {
			return err
		}

		err = s.dispatchIdentityCreatedEvent(ctx, info)
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
	Code       string
}

type ResumeAddIdentityPhoneOutput struct {
	IdentityInfo *identity.Info
}

func (s *Service) ResumeAddIdentityPhone(ctx context.Context, resolvedSession session.ResolvedSession, tokenString string, input *ResumeAddIdentityPhoneInput) (output *ResumeAddIdentityPhoneOutput, err error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID
	token, err := s.Store.GetToken(ctx, tokenString)
	defer func() {
		if err == nil {
			_, err = s.Store.ConsumeToken(ctx, tokenString)
		}
	}()

	if err != nil {
		return
	}
	err = token.CheckUser(userID)
	if err != nil {
		return
	}

	err = s.VerifyOTP(ctx, userID, model.AuthenticatorOOBChannelSMS, token.Identity.PhoneNumber, input.Code, false)
	if err != nil {
		return
	}

	spec, err := s.makeLoginIDSpec(input.LoginIDKey, token.Identity.PhoneNumber)
	if err != nil {
		return
	}

	var info *identity.Info
	err = s.Database.WithTx(ctx, func(ctx context.Context) error {
		info, err = s.prepareNewIdentity(ctx, userID, spec)
		if err != nil {
			return err
		}

		err = s.createIdentity(ctx, info)
		if err != nil {
			return err
		}

		claimName, ok := model.GetLoginIDKeyTypeClaim(info.LoginID.LoginIDType)
		if !ok {
			panic(fmt.Errorf("accountmanagement: unexpected login ID key"))
		}
		err = s.markClaimVerified(ctx, userID, claimName, info.LoginID.LoginID)
		if err != nil {
			return err
		}

		err = s.dispatchIdentityCreatedEvent(ctx, info)
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
	Channel    model.AuthenticatorOOBChannel
	IdentityID string
	LoginID    string
	LoginIDKey string
}

type StartUpdateIdentityPhoneOutput struct {
	IdentityInfo     *identity.Info
	NeedVerification bool
	Token            string
}

func (s *Service) StartUpdateIdentityPhone(ctx context.Context, resolvedSession session.ResolvedSession, input *StartUpdateIdentityPhoneInput) (*StartUpdateIdentityPhoneOutput, error) {
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
	err = s.Database.WithTx(ctx, func(ctx context.Context) error {
		oldInfo, newInfo, err := s.prepareUpdateIdentity(ctx, userID, input.IdentityID, spec)
		if err != nil {
			return err
		}
		info = newInfo

		verified, err := s.CheckIdentityVerified(ctx, newInfo)
		if err != nil {
			return err
		}
		needVerification = !verified && *s.Config.Verification.Claims.PhoneNumber.Enabled && *s.Config.Verification.Claims.PhoneNumber.Required
		if needVerification {
			target := info.LoginID.LoginID
			err = s.sendOTPCode(ctx, userID, input.Channel, target, false)
			if err != nil {
				return err
			}
			token, err = s.Store.GenerateToken(ctx, GenerateTokenOptions{
				UserID:              userID,
				IdentityChannel:     input.Channel,
				IdentityPhoneNumber: info.LoginID.LoginID,
				IdentityID:          info.ID,
			})
			if err != nil {
				return err
			}
			return nil
		}

		err = s.updateIdentity(ctx, oldInfo, newInfo)
		if err != nil {
			return err
		}

		err = s.dispatchIdentityCreatedEvent(ctx, newInfo)
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
	Code       string
}

type ResumeUpdateIdentityPhoneOutput struct {
	OldInfo *identity.Info
	NewInfo *identity.Info
}

func (s *Service) ResumeUpdateIdentityPhone(ctx context.Context, resolvedSession session.ResolvedSession, tokenString string, input *ResumeUpdateIdentityPhoneInput) (output *ResumeUpdateIdentityPhoneOutput, err error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID
	token, err := s.Store.GetToken(ctx, tokenString)
	defer func() {
		if err == nil {
			_, err = s.Store.ConsumeToken(ctx, tokenString)
		}
	}()

	if err != nil {
		return
	}
	err = token.CheckUser(userID)
	if err != nil {
		return
	}

	err = s.VerifyOTP(ctx, userID, model.AuthenticatorOOBChannelSMS, token.Identity.PhoneNumber, input.Code, false)
	if err != nil {
		return
	}

	spec, err := s.makeLoginIDSpec(input.LoginIDKey, token.Identity.PhoneNumber)
	if err != nil {
		return
	}

	var oldInfo *identity.Info
	var newInfo *identity.Info
	err = s.Database.WithTx(ctx, func(ctx context.Context) error {
		oldInfo, newInfo, err := s.prepareUpdateIdentity(ctx, userID, token.Identity.IdentityID, spec)
		if err != nil {
			return err
		}

		err = s.updateIdentity(ctx, oldInfo, newInfo)
		if err != nil {
			return err
		}

		claimName, ok := model.GetLoginIDKeyTypeClaim(newInfo.LoginID.LoginIDType)
		if !ok {
			panic(fmt.Errorf("accountmanagement: unexpected login ID key"))
		}
		err = s.markClaimVerified(ctx, userID, claimName, newInfo.LoginID.LoginID)
		if err != nil {
			return err
		}

		err = s.dispatchIdentityUpdatedEvent(ctx, oldInfo, newInfo)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	output = &ResumeUpdateIdentityPhoneOutput{
		OldInfo: oldInfo,
		NewInfo: newInfo,
	}
	return
}

type ResumeAddOrUpdateIdentityPhoneInput struct {
	LoginIDKey string
	Code       string
}

type ResumeAddOrUpdateIdentityPhoneOutput struct {
	OldInfo *identity.Info
	NewInfo *identity.Info
}

func (s *Service) ResumeAddOrUpdateIdentityPhone(ctx context.Context, resolvedSession session.ResolvedSession, tokenString string, input *ResumeAddOrUpdateIdentityPhoneInput) (*ResumeAddOrUpdateIdentityPhoneOutput, error) {
	token, err := s.Store.GetToken(ctx, tokenString)
	if err != nil {
		return nil, err
	}

	if token.Identity.IdentityID == "" {
		output, err := s.ResumeAddIdentityPhone(ctx, resolvedSession, tokenString, &ResumeAddIdentityPhoneInput{
			LoginIDKey: input.LoginIDKey,
			Code:       input.Code,
		})
		if err != nil {
			return nil, err
		}
		return &ResumeAddOrUpdateIdentityPhoneOutput{
			NewInfo: output.IdentityInfo,
		}, nil
	}

	output, err := s.ResumeUpdateIdentityPhone(ctx, resolvedSession, tokenString, &ResumeUpdateIdentityPhoneInput{
		LoginIDKey: input.LoginIDKey,
		Code:       input.Code,
	})
	if err != nil {
		return nil, err
	}
	return &ResumeAddOrUpdateIdentityPhoneOutput{
		OldInfo: output.OldInfo,
		NewInfo: output.NewInfo,
	}, nil
}

type DeleteIdentityPhoneInput struct {
	IdentityID string
}

type DeleteIdentityPhoneOutput struct {
	IdentityInfo *identity.Info
}

func (s *Service) DeleteIdentityPhone(ctx context.Context, resolvedSession session.ResolvedSession, input *DeleteIdentityPhoneInput) (*DeleteIdentityPhoneOutput, error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID
	identityID := input.IdentityID

	var info *identity.Info
	err := s.Database.WithTx(ctx, func(ctx context.Context) (err error) {
		info, err = s.prepareDeleteIdentity(ctx, userID, identityID)
		if err != nil {
			return err
		}

		err = s.deleteIdentity(ctx, info)
		if err != nil {
			return err
		}

		err = s.dispatchIdentityDeletedEvent(ctx, info)
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

func (s *Service) AddPasskey(ctx context.Context, resolvedSession session.ResolvedSession, input *AddPasskeyInput) (*AddPasskeyOutput, error) {
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
	authenticatorInfo, err := s.Authenticators.NewWithAuthenticatorID(ctx, authenticatorID, authenticatorSpec)
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
	err = s.Database.WithTx(ctx, func(ctx context.Context) error {
		identityInfo, err = s.prepareNewIdentity(ctx, userID, identitySpec)
		if err != nil {
			return err
		}

		err = s.createIdentity(ctx, identityInfo)
		if err != nil {
			return err
		}

		err = s.dispatchIdentityCreatedEvent(ctx, identityInfo)
		if err != nil {
			return err
		}

		err = s.Authenticators.Create(ctx, authenticatorInfo, false)
		if err != nil {
			return err
		}
		err = s.PasskeyService.ConsumeAttestationResponse(ctx, creationResponseBytes)
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

func (s *Service) DeletePasskey(ctx context.Context, resolvedSession session.ResolvedSession, input *DeletePasskeyInput) (*DeletePasskeyOutput, error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID
	identityID := input.IdentityID

	var info *identity.Info
	err := s.Database.WithTx(ctx, func(ctx context.Context) (err error) {
		info, err = s.prepareDeleteIdentity(ctx, userID, identityID)
		if err != nil {
			return err
		}

		err = s.deleteIdentity(ctx, info)
		if err != nil {
			return err
		}

		err = s.dispatchIdentityDeletedEvent(ctx, info)
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

func (s *Service) DeleteIdentityBiometric(ctx context.Context, resolvedSession session.ResolvedSession, input *DeleteIdentityBiometricInput) (*DeleteIdentityBiometricOuput, error) {
	identityID := input.IdentityID
	userID := resolvedSession.GetAuthenticationInfo().UserID

	var info *identity.Info
	err := s.Database.WithTx(ctx, func(ctx context.Context) (err error) {
		info, err = s.prepareDeleteIdentity(ctx, userID, identityID)
		if err != nil {
			return err
		}

		err = s.deleteIdentity(ctx, info)
		if err != nil {
			return err
		}

		err = s.dispatchIdentityDeletedEvent(ctx, info)
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
			Value: stringutil.NewUserInputString(loginID),
		},
	}
	return identitySpec, nil
}

func (s *Service) prepareNewIdentity(ctx context.Context, userID string, identitySpec *identity.Spec) (*identity.Info, error) {
	info, err := s.Identities.New(ctx, userID, identitySpec, identity.NewIdentityOptions{LoginIDEmailByPassBlocklistAllowlist: false})
	if err != nil {
		return nil, err
	}

	createDisabled := info.CreateDisabled(s.Config.Identity)
	if createDisabled {
		return nil, api.ErrIdentityModifyDisabled
	}

	if _, err := s.Identities.CheckDuplicated(ctx, info); err != nil {
		if identity.IsErrDuplicatedIdentity(err) {
			return nil, NewErrAccountManagementDuplicatedIdentity(err)
		}
		return nil, err
	}

	return info, nil
}

func (s *Service) prepareUpdateIdentity(ctx context.Context, userID string, identityID string, identitySpec *identity.Spec) (*identity.Info, *identity.Info, error) {
	oldInfo, err := s.Identities.Get(ctx, identityID)
	if err != nil {
		return nil, nil, err
	}

	if oldInfo.UserID != userID {
		return nil, nil, ErrAccountManagementIdentityNotOwnedbyToUser
	}

	newInfo, err := s.Identities.UpdateWithSpec(ctx, oldInfo, identitySpec, identity.NewIdentityOptions{
		LoginIDEmailByPassBlocklistAllowlist: false,
	})
	if err != nil {
		return nil, nil, err
	}

	updateDisabled := oldInfo.UpdateDisabled(s.Config.Identity)
	if updateDisabled {
		return nil, nil, api.ErrIdentityModifyDisabled
	}

	if _, err := s.Identities.CheckDuplicated(ctx, newInfo); err != nil {
		if identity.IsErrDuplicatedIdentity(err) {
			return nil, nil, NewErrAccountManagementDuplicatedIdentity(err)
		}
		return nil, nil, err
	}

	return oldInfo, newInfo, nil
}

func (s *Service) prepareDeleteIdentity(ctx context.Context, userID string, identityID string) (*identity.Info, error) {
	info, err := s.Identities.Get(ctx, identityID)
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

func (s *Service) CheckIdentityVerified(ctx context.Context, info *identity.Info) (bool, error) {
	claims, err := s.Verification.GetIdentityVerificationStatus(ctx, info)
	if err != nil {
		return false, err
	}
	if len(claims) == 0 {
		return false, nil
	}
	claim := claims[0]
	return claim.Verified, nil
}

func (s *Service) createIdentity(ctx context.Context, info *identity.Info) error {
	return s.Identities.Create(ctx, info)
}

func (s *Service) updateIdentity(ctx context.Context, oldInfo *identity.Info, newInfo *identity.Info) error {
	return s.Identities.Update(ctx, oldInfo, newInfo)
}

func (s *Service) deleteIdentity(ctx context.Context, info *identity.Info) error {
	return s.Identities.Delete(ctx, info)
}

func (s *Service) dispatchIdentityCreatedEvent(ctx context.Context, info *identity.Info) (err error) {
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
		err = s.Events.DispatchEventOnCommit(ctx, e)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) dispatchIdentityUpdatedEvent(ctx context.Context, oldInfo *identity.Info, newInfo *identity.Info) (err error) {
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
		err = s.Events.DispatchEventOnCommit(ctx, e)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) dispatchIdentityDeletedEvent(ctx context.Context, info *identity.Info) (err error) {
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
		err = s.Events.DispatchEventOnCommit(ctx, e)
		if err != nil {
			return err
		}
	}

	return nil
}

type StartAddIdentityOAuthInput struct {
	Alias       string
	RedirectURI string
}

type StartAddIdentityOAuthOutput struct {
	Token            string
	AuthorizationURL string
}

func (s *Service) StartAddIdentityOAuth(ctx context.Context, resolvedSession session.ResolvedSession, input *StartAddIdentityOAuthInput) (*StartAddIdentityOAuthOutput, error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID

	var err error
	var token string
	var authorizationURL string
	err = s.Database.WithTx(ctx, func(ctx context.Context) error {
		output, err := s.StartAdding(ctx, &StartAddingInput{
			UserID:      userID,
			Alias:       input.Alias,
			RedirectURI: input.RedirectURI,
			IncludeStateAuthorizationURLAndBindStateToToken: false,
		})
		if err != nil {
			return err
		}

		token = output.Token
		authorizationURL = output.AuthorizationURL

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &StartAddIdentityOAuthOutput{
		Token:            token,
		AuthorizationURL: authorizationURL,
	}, nil
}

type FinishAddingIdentityOAuthInput struct {
	Token string
	Query string
}

type FinishAddingIdentityOAuthOutput struct {
}

func (s *Service) FinishAddingIdentityOAuth(ctx context.Context, resolvedSession session.ResolvedSession, input *FinishAddingIdentityOAuthInput) (*FinishAddingIdentityOAuthOutput, error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID

	err := s.Database.WithTx(ctx, func(ctx context.Context) error {
		_, err := s.FinishAdding(ctx, &FinishAddingInput{
			UserID: userID,
			Token:  input.Token,
			Query:  input.Query,
		})
		if err != nil {
			if identity.IsErrDuplicatedIdentity(err) {
				return NewErrAccountManagementDuplicatedIdentity(err)
			}

			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &FinishAddingIdentityOAuthOutput{}, nil
}

type DeleteIdentityOAuthInput struct {
	IdentityID string
}

type DeleteIdentityOAuthOutput struct {
	IdentityInfo *identity.Info
}

func (s *Service) DeleteIdentityOAuth(ctx context.Context, resolvedSession session.ResolvedSession, input *DeleteIdentityOAuthInput) (*DeleteIdentityOAuthOutput, error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID
	identityID := input.IdentityID

	var info *identity.Info
	err := s.Database.WithTx(ctx, func(ctx context.Context) (err error) {
		info, err = s.prepareDeleteIdentity(ctx, userID, identityID)
		if err != nil {
			return err
		}

		err = s.deleteIdentity(ctx, info)
		if err != nil {
			return err
		}

		err = s.dispatchIdentityDeletedEvent(ctx, info)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &DeleteIdentityOAuthOutput{IdentityInfo: info}, nil
}

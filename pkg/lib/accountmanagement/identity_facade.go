package accountmanagement

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/lib/session"
)

type IdentityFacade struct {
	Config         *config.AppConfig
	Store          Store
	Identities     IdentityService
	Events         EventService
	OTPSender      OTPSender
	OTPCodeService OTPCodeService
	Verification   VerificationService
}

func (i *IdentityFacade) MakeLoginIDSpec(loginIDKey string, loginID string) (*identity.Spec, error) {
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
	isVerified := false
	identityInfo, err := i.Identities.New(userID, identitySpec, identity.NewIdentityOptions{LoginIDEmailByPassBlocklistAllowlist: false})
	if err != nil {
		return nil, isVerified, err
	}
	createDisabled := identityInfo.CreateDisabled(i.Config.Identity)
	if createDisabled {
		return nil, isVerified, api.ErrIdentityModifyDisabled
	}

	if _, err := i.Identities.CheckDuplicated(identityInfo); err != nil {
		return nil, isVerified, err
	}

	if needVerify {
		claims, err := i.Verification.GetIdentityVerificationStatus(identityInfo)
		if err != nil {
			return nil, isVerified, err
		}
		// if verified, create identity immediately
		if len(claims) > 0 && claims[0].Verified {
			isVerified = true
			if err = i.Identities.Create(identityInfo); err != nil {
				if identity.IsErrDuplicatedIdentity(err) {
					return identityInfo, isVerified, nil
				}
				return nil, isVerified, err
			}
			if err = i.dispatchIdentityCreatedEvent(identityInfo); err != nil {
				return nil, isVerified, err
			}
		}
	}

	return identityInfo, isVerified, nil
}

func (i *IdentityFacade) UpdateIdentity(userID string, identityID string, identitySpec *identity.Spec, needVerify bool) (*identity.Info, bool, error) {
	isVerified := false
	oldInfo, err := i.Identities.Get(identityID)
	fmt.Printf("oldInfo: %v\n", oldInfo)
	if err != nil {
		return nil, isVerified, err
	}

	if oldInfo.UserID != userID {
		return nil, isVerified, ErrAccountManagementIdentityNotOwnedbyToUser
	}

	newInfo, err := i.Identities.UpdateWithSpec(oldInfo, identitySpec, identity.NewIdentityOptions{
		LoginIDEmailByPassBlocklistAllowlist: false,
	})
	if err != nil {
		return nil, isVerified, err
	}

	updateDisabled := oldInfo.UpdateDisabled(i.Config.Identity)
	if updateDisabled {
		return nil, isVerified, api.ErrIdentityModifyDisabled
	}

	if _, err := i.Identities.CheckDuplicated(newInfo); err != nil {
		if identity.IsErrDuplicatedIdentity(err) {
			s1 := oldInfo.ToSpec()
			s2 := newInfo.ToSpec()
			return nil, isVerified, identity.NewErrDuplicatedIdentity(&s2, &s1)
		}
		return nil, isVerified, err
	}

	if needVerify {
		claims, err := i.Verification.GetIdentityVerificationStatus(newInfo)
		if err != nil {
			return nil, isVerified, err
		}
		// if verified, update identity immediately
		if len(claims) > 0 && claims[0].Verified {
			isVerified = true
			if err := i.Identities.Update(oldInfo, newInfo); err != nil {
				return nil, isVerified, err
			}

			if err = i.dispatchIdentityUpdatedEvent(oldInfo, newInfo); err != nil {
				return nil, isVerified, err
			}
		}
	}

	return newInfo, isVerified, nil
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

	deleteDiabled := identityInfo.DeleteDisabled(i.Config.Identity)
	if deleteDiabled {
		return nil, api.ErrIdentityModifyDisabled
	}

	if err := i.Identities.Delete(identityInfo); err != nil {
		return nil, err
	}

	if err := i.dispatchIdentityRemovedEvent(identityInfo); err != nil {
		return nil, err
	}

	return identityInfo, nil
}

func (i *IdentityFacade) VerifyIdentity(input *verifyIdentityInput) (verifiedClaim *verification.Claim, err error) {
	var loginIDValue string
	var loginIDType model.LoginIDKeyType
	token := input.Token
	switch {
	case token.IdentityToken.Email != "":
		loginIDValue = token.IdentityToken.Email
		loginIDType = model.LoginIDKeyTypeEmail
	case token.IdentityToken.PhoneNumber != "":
		loginIDValue = token.IdentityToken.PhoneNumber
		loginIDType = model.LoginIDKeyTypePhone
	default:
		return nil, ErrAccountManagementTokenInvalid
	}

	err = i.OTPCodeService.VerifyOTP(
		otp.KindVerification(i.Config, input.Channel),
		loginIDValue,
		input.Code,
		&otp.VerifyOptions{UserID: input.UserID},
	)
	if apierrors.IsKind(err, otp.InvalidOTPCode) {
		return nil, verification.ErrInvalidVerificationCode
	} else if err != nil {
		return nil, err
	}

	var claimName model.ClaimName
	claimName, ok := model.GetLoginIDKeyTypeClaim(loginIDType)
	if !ok {
		panic(fmt.Errorf("accountmanagement: unexpected login ID key"))
	}

	verifiedClaim = i.Verification.NewVerifiedClaim(input.UserID, string(claimName), loginIDValue)

	err = i.Verification.MarkClaimVerified(verifiedClaim)
	if err != nil {
		return nil, err
	}

	return verifiedClaim, nil
}

func (i *IdentityFacade) SendOTPCode(input *sendOTPCodeInput) error {
	var msgType otp.MessageType
	switch input.Channel {
	case model.AuthenticatorOOBChannelWhatsapp:
		msgType = otp.MessageTypeWhatsappCode
	case model.AuthenticatorOOBChannelSMS:
		msgType = otp.MessageTypeVerification
	case model.AuthenticatorOOBChannelEmail:
		msgType = otp.MessageTypeVerification
	default:
		panic(fmt.Errorf("accountmanagement: unknown channel"))
	}

	msg, err := i.OTPSender.Prepare(input.Channel, input.Target, otp.FormCode, msgType)
	if !input.isResend && apierrors.IsKind(err, ratelimit.RateLimited) {
		return nil
	} else if err != nil {
		return err
	}
	defer msg.Close()

	code, err := i.OTPCodeService.GenerateOTP(
		otp.KindVerification(i.Config, input.Channel),
		input.Target,
		otp.FormCode,
		&otp.GenerateOptions{},
	)
	// If it is not resend (switch between page), we should not send and return rate limit error to the caller.
	if !input.isResend && apierrors.IsKind(err, ratelimit.RateLimited) {
		return nil
	} else if err != nil {
		return err
	}

	err = i.OTPSender.Send(msg, otp.SendOptions{OTP: code})
	if err != nil {
		return err
	}

	return nil
}

func (i *IdentityFacade) StartIdentityWithVerification(resolvedSession session.ResolvedSession, input *startIdentityWithVerificationInput) (output *StartIdentityWithVerificationOutput, err error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID

	var newInfo *identity.Info

	// Currently only LoginID requires verification.
	identitySpec := input.IdentitySpec

	var isVerified bool
	switch {
	case input.isUpdate:
		newInfo, isVerified, err = i.UpdateIdentity(userID, input.IdentityID, identitySpec, true)
	case !input.isUpdate:
		newInfo, isVerified, err = i.CreateIdentity(userID, identitySpec, true)
	}
	if err != nil {
		return nil, err
	}

	if !isVerified {
		err = i.SendOTPCode(&sendOTPCodeInput{
			Channel:  input.Channel,
			Target:   input.LoginID,
			isResend: false,
		})
		if err != nil {
			return nil, err
		}
	}

	return &StartIdentityWithVerificationOutput{
		IdentityInfo:     newInfo,
		NeedVerification: !isVerified,
	}, nil
}

func (i *IdentityFacade) CreateIdentityWithVerification(resolvedSession session.ResolvedSession, input *CreateIdentityWithVerificationInput) (*CreateIdentityWithVerificationOutput, error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID

	identitySpec := input.IdentitySpec

	var identityInfo *identity.Info
	verifiedClaim, err := i.VerifyIdentity(&verifyIdentityInput{
		UserID:  userID,
		Token:   input.Token,
		Channel: input.Channel,
		Code:    input.Code,
	})
	if err != nil {
		return nil, err
	}

	// Create identity after verification
	identityInfo, _, err = i.CreateIdentity(userID, identitySpec, false)
	if err != nil {
		return nil, err
	}

	err = i.dispatchIdentityVerifiedEvent(identityInfo, verifiedClaim)
	if err != nil {
		return nil, err
	}

	return &CreateIdentityWithVerificationOutput{IdentityInfo: identityInfo}, nil
}

func (i *IdentityFacade) UpdateIdentityWithVerification(resolvedSession session.ResolvedSession, input *UpdateIdentityWithVerificationInput) (*UpdateIdentityWithVerificationOutput, error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID
	identityID := input.IdentityID

	identitySpec := input.IdentitySpec

	var identityInfo *identity.Info
	verifiedClaim, err := i.VerifyIdentity(&verifyIdentityInput{
		UserID:  userID,
		Token:   input.Token,
		Channel: input.Channel,
		Code:    input.Code,
	})
	if err != nil {
		return nil, err
	}

	// Update identity after verification
	identityInfo, _, err = i.UpdateIdentity(userID, identityID, identitySpec, false)
	if err != nil {
		return nil, err
	}

	err = i.dispatchIdentityVerifiedEvent(identityInfo, verifiedClaim)
	if err != nil {
		return nil, err
	}

	return &UpdateIdentityWithVerificationOutput{IdentityInfo: identityInfo}, nil
}

func (i *IdentityFacade) dispatchIdentityCreatedEvent(identityInfo *identity.Info) (err error) {
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

func (i *IdentityFacade) dispatchIdentityUpdatedEvent(identityAfterUpdate *identity.Info, identityBeforeUpdate *identity.Info) (err error) {
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

func (i *IdentityFacade) dispatchIdentityRemovedEvent(identityInfo *identity.Info) (err error) {
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

func (i *IdentityFacade) dispatchIdentityVerifiedEvent(identityInfo *identity.Info, verifiedClaim *verification.Claim) error {
	var e event.Payload
	if payload, ok := nonblocking.NewIdentityVerifiedEventPayload(
		model.UserRef{
			Meta: model.Meta{
				ID: identityInfo.UserID,
			},
		},
		identityInfo.ToModel(),
		string(verifiedClaim.Name),
		false,
	); ok {
		e = payload
	}

	if e != nil {
		if err := i.Events.DispatchEventOnCommit(e); err != nil {
			return err
		}
	}
	return nil
}

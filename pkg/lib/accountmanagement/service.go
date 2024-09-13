package accountmanagement

import (
	"encoding/json"
	"errors"
	"fmt"

	"time"

	"github.com/authgear/oauthrelyingparty/pkg/api/oauthrelyingparty"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator/service"
	"github.com/authgear/authgear-server/pkg/lib/authn/challenge"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity/biometric"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/facade"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
	"github.com/authgear/authgear-server/pkg/util/deviceinfo"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

type ChallengeProvider interface {
	Consume(token string) (*challenge.Purpose, error)
}

type UserService interface {
	Get(id string, role accesscontrol.Role) (*model.User, error)
	UpdateMFAEnrollment(userID string, t *time.Time) error
}

type Store interface {
	GenerateToken(options GenerateTokenOptions) (string, error)
	GetToken(tokenStr string) (*Token, error)
	ConsumeToken(tokenStr string) (*Token, error)
}

type OAuthProvider interface {
	GetProviderConfig(alias string) (oauthrelyingparty.ProviderConfig, error)
	GetAuthorizationURL(alias string, options oauthrelyingparty.GetAuthorizationURLOptions) (string, error)
	GetUserProfile(alias string, options oauthrelyingparty.GetUserProfileOptions) (oauthrelyingparty.UserProfile, error)
}

type BiometricIdentityProvider interface {
	ParseRequestUnverified(requestJWT string) (*biometric.Request, error)
}

type IdentityService interface {
	Get(id string) (*identity.Info, error)
	ListByUser(userID string) ([]*identity.Info, error)
	New(userID string, spec *identity.Spec, options identity.NewIdentityOptions) (*identity.Info, error)
	UpdateWithSpec(is *identity.Info, spec *identity.Spec, options identity.NewIdentityOptions) (*identity.Info, error)
	CheckDuplicated(info *identity.Info) (dupe *identity.Info, err error)
	Create(info *identity.Info) error
	Update(oldInfo *identity.Info, newInfo *identity.Info) error
	Delete(is *identity.Info) error
}

type IdentityAction interface {
	CreateIdentity(userID string, identitySpec *identity.Spec, needVerify bool) (*identity.Info, bool, error)
	UpdateIdentity(userID string, identityID string, identitySpec *identity.Spec, needVerify bool) (*identity.Info, bool, error)
	RemoveIdentity(userID string, identityID string) (*identity.Info, error)
	MakeLoginIDSpec(loginIDKey string, loginID string) (*identity.Spec, error)
}

type EventService interface {
	DispatchEventOnCommit(payload event.Payload) error
}

type AuthenticatorService interface {
	NewWithAuthenticatorID(authenticatorID string, spec *authenticator.Spec) (*authenticator.Info, error)
	List(userID string, filters ...authenticator.Filter) ([]*authenticator.Info, error)
	Create(authenticatorInfo *authenticator.Info, markVerified bool) error
	Update(authenticatorInfo *authenticator.Info) error
	UpdatePassword(authenticatorInfo *authenticator.Info, options *service.UpdatePasswordOptions) (changed bool, info *authenticator.Info, err error)
	VerifyWithSpec(info *authenticator.Info, spec *authenticator.Spec, options *facade.VerifyOptions) (verifyResult *service.VerifyResult, err error)
}

type AuthenticationInfoService interface {
	Save(entry *authenticationinfo.Entry) error
}

type PasskeyService interface {
	ConsumeAttestationResponse(attestationResponse []byte) (err error)
}

type SettingsDeleteAccountSuccessUIInfoResolver interface {
	SetAuthenticationInfoInQuery(redirectURI string, e *authenticationinfo.Entry) string
}

type OTPSender interface {
	Prepare(channel model.AuthenticatorOOBChannel, target string, form otp.Form, typ otp.MessageType) (*otp.PreparedMessage, error)
	Send(msg *otp.PreparedMessage, opts otp.SendOptions) error
}

type OTPCodeService interface {
	GenerateOTP(kind otp.Kind, target string, form otp.Form, opt *otp.GenerateOptions) (string, error)
	VerifyOTP(kind otp.Kind, target string, otp string, opts *otp.VerifyOptions) error
}

type VerificationService interface {
	NewVerifiedClaim(userID string, claimName string, claimValue string) *verification.Claim
	MarkClaimVerified(claim *verification.Claim) error
	GetIdentityVerificationStatus(i *identity.Info) ([]verification.ClaimStatus, error)
}

type Service struct {
	Database                  *appdb.Handle
	Config                    *config.AppConfig
	Challenges                ChallengeProvider
	Users                     UserService
	Store                     Store
	OAuthProvider             OAuthProvider
	BiometricProvider         BiometricIdentityProvider
	Identities                IdentityService
	IdentityAction            IdentityAction
	Events                    EventService
	Authenticators            AuthenticatorService
	AuthenticationInfoService AuthenticationInfoService
	PasskeyService            PasskeyService
	UIInfoResolver            SettingsDeleteAccountSuccessUIInfoResolver
	OTPSender                 OTPSender
	OTPCodeService            OTPCodeService
	Verification              VerificationService
}

func (s *Service) verifyIdentity(input *verifyIdentityInput) (verifiedClaim *verification.Claim, err error) {
	defer func() {
		if err == nil {
			_, err = s.Store.ConsumeToken(input.Token)
		}
	}()

	token, err := s.Store.GetToken(input.Token)
	err = token.CheckUser(input.UserID)
	if err != nil {
		return nil, err
	}

	var loginIDValue string
	var loginIDType model.LoginIDKeyType
	switch {
	case token.Email != "":
		loginIDValue = token.Email
		loginIDType = model.LoginIDKeyTypeEmail
	case token.PhoneNumber != "":
		loginIDValue = token.PhoneNumber
		loginIDType = model.LoginIDKeyTypePhone
	default:
		return nil, ErrAccountManagementTokenInvalid
	}

	err = s.OTPCodeService.VerifyOTP(
		otp.KindVerification(s.Config, input.Channel),
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

	verifiedClaim = s.Verification.NewVerifiedClaim(input.UserID, string(claimName), loginIDValue)

	// NodeDoVerifyIdentity GetEffects()
	err = s.Verification.MarkClaimVerified(verifiedClaim)
	if err != nil {
		return nil, err
	}

	return verifiedClaim, nil
}

func (s *Service) sendOTPCode(input *sendOTPCodeInput) error {
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

	msg, err := s.OTPSender.Prepare(input.Channel, input.Target, otp.FormCode, msgType)
	if !input.isResend && apierrors.IsKind(err, ratelimit.RateLimited) {
		return nil
	} else if err != nil {
		return err
	}
	defer msg.Close()

	code, err := s.OTPCodeService.GenerateOTP(
		otp.KindVerification(s.Config, input.Channel),
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

	err = s.OTPSender.Send(msg, otp.SendOptions{OTP: code})
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) StartAdding(input *StartAddingInput) (*StartAddingOutput, error) {
	state := ""
	if input.IncludeStateAuthorizationURLAndBindStateToToken {
		state = GenerateRandomState()
	}

	param := oauthrelyingparty.GetAuthorizationURLOptions{
		RedirectURI: input.RedirectURI,
		State:       state,
	}

	authorizationURL, err := s.OAuthProvider.GetAuthorizationURL(input.Alias, param)
	if err != nil {
		return nil, err
	}

	token, err := s.Store.GenerateToken(GenerateTokenOptions{
		UserID:      input.UserID,
		Alias:       input.Alias,
		RedirectURI: input.RedirectURI,
		MaybeState:  state,
	})
	if err != nil {
		return nil, err
	}

	return &StartAddingOutput{
		Token:            token,
		AuthorizationURL: authorizationURL,
	}, nil
}

func (s *Service) FinishAdding(input *FinishAddingInput) (*FinishAddingOutput, error) {
	token, err := s.Store.ConsumeToken(input.Token)
	if err != nil {
		if errors.Is(err, ErrAccountManagementTokenInvalid) {
			return nil, ErrOAuthTokenInvalid
		}
		return nil, err
	}

	err = token.CheckUser(input.UserID)
	if err != nil {
		if errors.Is(err, ErrAccountManagementTokenNotBoundToUser) {
			return nil, ErrOAuthTokenNotBoundToUser
		}
		return nil, err
	}

	state, err := ExtractStateFromQuery(input.Query)
	if err != nil {
		return nil, err
	}

	err = token.CheckState(state)
	if err != nil {
		return nil, err
	}

	providerConfig, err := s.OAuthProvider.GetProviderConfig(token.Alias)
	if err != nil {
		return nil, err
	}

	emptyNonce := ""
	userProfile, err := s.OAuthProvider.GetUserProfile(token.Alias, oauthrelyingparty.GetUserProfileOptions{
		Query:       input.Query,
		RedirectURI: token.RedirectURI,
		Nonce:       emptyNonce,
	})
	if err != nil {
		return nil, err
	}

	providerID := providerConfig.ProviderID()
	spec := &identity.Spec{
		Type: model.IdentityTypeOAuth,
		OAuth: &identity.OAuthSpec{
			ProviderID:     providerID,
			SubjectID:      userProfile.ProviderUserID,
			RawProfile:     userProfile.ProviderRawProfile,
			StandardClaims: userProfile.StandardAttributes,
		},
	}

	info, err := s.Identities.New(
		token.UserID,
		spec,
		// We are not adding Login ID here so the options is irrelevant.
		identity.NewIdentityOptions{},
	)
	if err != nil {
		return nil, err
	}

	err = s.Database.WithTx(func() error {
		_, err = s.Identities.CheckDuplicated(info)
		if err != nil {
			return err
		}

		err = s.Identities.Create(info)
		if err != nil {
			return err
		}

		evt := &nonblocking.IdentityOAuthConnectedEventPayload{
			UserRef: model.UserRef{
				Meta: model.Meta{
					ID: info.UserID,
				},
			},
			Identity: info.ToModel(),
			AdminAPI: false,
		}

		err = s.Events.DispatchEventOnCommit(evt)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &FinishAddingOutput{}, nil
}

func (s *Service) StartCreateIdentityWithVerification(resolvedSession session.ResolvedSession, input *StartCreateIdentityWithVerificationInput) (output *StartIdentityWithVerificationOutput, err error) {
	return s.startIdentityWithVerification(resolvedSession, &startIdentityWithVerificationInput{
		LoginID:    input.LoginID,
		LoginIDKey: input.LoginIDKey,
		Channel:    input.Channel,
		isUpdate:   false,
	})
}

func (s *Service) StartUpdateIdentityWithVerification(resolvedSession session.ResolvedSession, input *StartUpdateIdentityWithVerificationInput) (output *StartIdentityWithVerificationOutput, err error) {
	return s.startIdentityWithVerification(resolvedSession, &startIdentityWithVerificationInput{
		LoginID:    input.LoginID,
		LoginIDKey: input.LoginIDKey,
		Channel:    input.Channel,
		IdentityID: input.IdentityID,
		isUpdate:   true,
	})
}

func (s *Service) startIdentityWithVerification(resolvedSession session.ResolvedSession, input *startIdentityWithVerificationInput) (output *StartIdentityWithVerificationOutput, err error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID
	var token string

	var newInfo *identity.Info

	// Currently only LoginID requires verification.
	identitySpec, err := s.IdentityAction.MakeLoginIDSpec(input.LoginIDKey, input.LoginID)

	var needVerify bool
	err = s.Database.WithTx(func() error {
		switch {
		case input.isUpdate:
			newInfo, needVerify, err = s.IdentityAction.UpdateIdentity(userID, input.IdentityID, identitySpec, true)
		case !input.isUpdate:
			newInfo, needVerify, err = s.IdentityAction.CreateIdentity(userID, identitySpec, true)
		}
		if err != nil {
			return err
		}

		if !needVerify {
			// Already Create / Update, we can skip send OTP code.
			return nil
		}

		loginIDType := newInfo.LoginID.LoginIDType
		switch loginIDType {
		case model.LoginIDKeyTypeEmail:
			token, err = s.Store.GenerateToken(GenerateTokenOptions{
				UserID:     userID,
				Email:      input.LoginID,
				IdentityID: input.IdentityID,
			})
		case model.LoginIDKeyTypePhone:
			token, err = s.Store.GenerateToken(GenerateTokenOptions{
				UserID:      userID,
				PhoneNumber: input.LoginID,
				IdentityID:  input.IdentityID,
			})
		}
		if err != nil {
			return err
		}

		err = s.sendOTPCode(&sendOTPCodeInput{
			Channel:  input.Channel,
			Target:   input.LoginID,
			isResend: false,
		})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &StartIdentityWithVerificationOutput{
		Token:            token,
		NeedVerification: needVerify,
	}, nil
}

func (s *Service) ResumeAddingIdentityWithVerification(resolvedSession session.ResolvedSession, input *ResumeAddingIdentityWithVerificationInput) (output *ResumeAddingIdentityWithVerificationOutput, err error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID
	token, err := s.Store.GetToken(input.Token)
	if err != nil {
		return nil, err
	}
	err = token.CheckUser(userID)
	if err != nil {
		return nil, err
	}

	var loginID string
	var loginIDKeyType model.LoginIDKeyType
	identityID := token.IdentityID

	switch {
	case token.Email != "":
		loginID = token.Email
		loginIDKeyType = model.LoginIDKeyTypeEmail
	case token.PhoneNumber != "":
		loginID = token.PhoneNumber
		loginIDKeyType = model.LoginIDKeyTypePhone
	default:
		return nil, ErrAccountManagementTokenInvalid
	}

	return &ResumeAddingIdentityWithVerificationOutput{
		Token:          input.Token,
		LoginID:        loginID,
		LoginIDKeyType: loginIDKeyType,
		IdentityID:     identityID,
	}, nil
}

func (s *Service) ResendOTPCode(input *ResendOTPCodeInput) (err error) {
	// Either it is a switch page or resend
	isResend := !input.isSwitchPage
	err = s.sendOTPCode(&sendOTPCodeInput{
		Channel:  input.Channel,
		Target:   input.LoginID,
		isResend: isResend,
	})
	if err != nil {
		return err
	}

	return nil
}

// If have OAuthSessionID, it means the user is changing password after login with SDK.
// Then do special handling such as authenticationInfo
func (s *Service) ChangePassword(resolvedSession session.ResolvedSession, input *ChangePasswordInput) (*ChangePasswordOutput, error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID

	ais, err := s.Authenticators.List(
		userID,
		authenticator.KeepType(model.AuthenticatorTypePassword),
		authenticator.KeepKind(authenticator.KindPrimary),
	)
	if err != nil {
		return &ChangePasswordOutput{}, err
	}
	if len(ais) == 0 {
		return &ChangePasswordOutput{}, api.ErrNoPassword
	}
	oldInfo := ais[0]
	_, err = s.Authenticators.VerifyWithSpec(oldInfo, &authenticator.Spec{
		Password: &authenticator.PasswordSpec{
			PlainPassword: input.OldPassword,
		},
	}, nil)
	if err != nil {
		err = api.ErrInvalidCredentials
		return &ChangePasswordOutput{}, err
	}
	changed, newInfo, err := s.Authenticators.UpdatePassword(oldInfo, &service.UpdatePasswordOptions{
		SetPassword:    true,
		PlainPassword:  input.NewPassword,
		SetExpireAfter: true,
	})
	if err != nil {
		return &ChangePasswordOutput{}, err
	}
	if changed {
		err = s.Database.WithTx(func() error {
			err = s.Authenticators.Update(newInfo)
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return &ChangePasswordOutput{}, err
		}

	}
	redirectURI := input.RedirectURI
	// If is changing password with SDK.
	if input.OAuthSessionID != "" {
		authInfo := resolvedSession.GetAuthenticationInfo()
		authenticationInfoEntry := authenticationinfo.NewEntry(authInfo, input.OAuthSessionID, "")

		err = s.Database.WithTx(func() error {
			err = s.AuthenticationInfoService.Save(authenticationInfoEntry)
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return &ChangePasswordOutput{}, err
		}
		redirectURI = s.UIInfoResolver.SetAuthenticationInfoInQuery(input.RedirectURI, authenticationInfoEntry)
	}
	return &ChangePasswordOutput{RedirectURI: redirectURI}, nil

}

func (s *Service) CreateAdditionalPassword(input CreateAdditionalPasswordInput) error {
	spec := &authenticator.Spec{
		UserID:    input.UserID,
		IsDefault: false,
		Kind:      model.AuthenticatorKindSecondary,
		Type:      model.AuthenticatorTypePassword,
		Password: &authenticator.PasswordSpec{
			PlainPassword: input.Password,
		},
	}
	info, err := s.Authenticators.NewWithAuthenticatorID(input.NewAuthenticatorID, spec)
	if err != nil {
		return err
	}
	return s.CreateAuthenticator(info)
}

func (s *Service) CreateAuthenticator(authenticatorInfo *authenticator.Info) error {
	err := s.Database.WithTx(func() error {
		err := s.Authenticators.Create(authenticatorInfo, false)
		if err != nil {
			return err
		}
		if authenticatorInfo.Kind == authenticator.KindSecondary {
			err = s.Users.UpdateMFAEnrollment(authenticatorInfo.UserID, nil)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) AddPasskey(resolvedSession session.ResolvedSession, input *AddPasskeyInput) (*AddPasskeyOutput, error) {
	// NodePromptCreatePasskey ReactTo
	// case inputNodePromptCreatePasskey.IsCreationResponse()
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
		identityInfo, _, err = s.IdentityAction.CreateIdentity(userID, identitySpec, false)
		if err != nil {
			return err
		}
		// NodeDoCreatePasskey GetEffects()
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

func (s *Service) RemovePasskey(resolvedSession session.ResolvedSession, input *RemovePasskeyInput) (*RemovePasskeyOutput, error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID
	identityID := input.IdentityID

	var identityInfo *identity.Info
	err := s.Database.WithTx(func() (err error) {
		// case *nodes.NodeDoUseUser: (Passkey skip DeleteDisabled check)
		identityInfo, err = s.IdentityAction.RemoveIdentity(userID, identityID)
		if err != nil {
			return err
		}
		// NodeDoRemoveIdentity GetEffects() -> EffectOnCommit()
		// Passkey no PayloadEvent for EffectOnCommit

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &RemovePasskeyOutput{IdentityInfo: identityInfo}, nil
}

func (s *Service) AddBiometric(resolvedSession session.ResolvedSession, input *AddBiometricInput) (*AddBiometricOutput, error) {

	// EdgeUseIdentityBiometric
	enabled := false
	for _, t := range s.Config.Authentication.Identities {
		if t == model.IdentityTypeBiometric {
			enabled = true
			break
		}
	}

	if !enabled {
		return nil, api.NewInvariantViolated(
			"BiometricDisallowed",
			"biometric is not allowed",
			nil,
		)
	}

	jwt := input.JWTToken

	request, err := s.BiometricProvider.ParseRequestUnverified(jwt)
	if err != nil {
		return nil, api.ErrInvalidCredentials
	}

	purpose, err := s.Challenges.Consume(request.Challenge)
	if err != nil || *purpose != challenge.PurposeBiometricRequest {
		return nil, api.ErrInvalidCredentials
	}

	// request.Action case: identitybiometric.RequestActionSetup
	displayName := deviceinfo.DeviceModel(request.DeviceInfo)
	if displayName == "" {
		return nil, api.ErrInvalidCredentials
	}
	if request.Key == nil {
		return nil, api.ErrInvalidCredentials
	}

	key, err := json.Marshal(request.Key)
	if err != nil {
		return nil, err
	}

	userID := resolvedSession.GetAuthenticationInfo().UserID

	// IsCreating: true
	identitySpec := &identity.Spec{
		Type: model.IdentityTypeBiometric,
		Biometric: &identity.BiometricSpec{
			KeyID:      request.KeyID,
			Key:        string(key),
			DeviceInfo: request.DeviceInfo,
		},
	}

	var identityInfo *identity.Info
	// EdgeDoCreateIdentity
	err = s.Database.WithTx(func() error {
		user, err := s.Users.Get(userID, accesscontrol.RoleGreatest)
		if err != nil {
			return err
		}

		if user.IsAnonymous {
			return api.NewInvariantViolated(
				"AnonymousUserAddIdentity",
				"anonymous user cannot add identity",
				nil,
			)
		}

		identityInfo, _, err = s.IdentityAction.CreateIdentity(userID, identitySpec, false)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &AddBiometricOutput{IdentityInfo: identityInfo}, nil
}

func (s *Service) RemoveBiometric(resolvedSession session.ResolvedSession, input *RemoveBiometricInput) (*RemoveBiometricOuput, error) {
	identityID := input.IdentityID
	userID := resolvedSession.GetAuthenticationInfo().UserID

	var identityInfo *identity.Info
	err := s.Database.WithTx(func() (err error) {
		// case *nodes.NodeDoUseUser: (Biometric skip DeleteDisabled check)
		identityInfo, err = s.IdentityAction.RemoveIdentity(userID, identityID)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &RemoveBiometricOuput{IdentityInfo: identityInfo}, nil
}

func (s *Service) AddUsername(resolvedSession session.ResolvedSession, input *AddUsernameInput) (*AddUsernameOutput, error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID
	loginKey := input.LoginIDKey
	loginID := input.LoginID

	identitySpec, err := s.IdentityAction.MakeLoginIDSpec(loginKey, loginID)

	var identityInfo *identity.Info
	err = s.Database.WithTx(func() error {
		identityInfo, _, err = s.IdentityAction.CreateIdentity(userID, identitySpec, false)
		if err != nil {
			return err
		}
		return nil

	})
	if err != nil {
		return nil, err
	}

	return &AddUsernameOutput{IdentityInfo: identityInfo}, nil

}

func (s *Service) UpdateUsername(resolvedSession session.ResolvedSession, input *UpdateUsernameInput) (*UpdateUsernameOutput, error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID
	loginKey := input.LoginIDKey
	loginID := input.LoginID
	identityID := input.IdentityID

	identitySpec, err := s.IdentityAction.MakeLoginIDSpec(loginKey, loginID)
	if err != nil {
		return nil, err
	}

	var identityInfo *identity.Info
	err = s.Database.WithTx(func() error {
		identityInfo, _, err = s.IdentityAction.UpdateIdentity(userID, identityID, identitySpec, false)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &UpdateUsernameOutput{IdentityInfo: identityInfo}, nil
}

func (s *Service) RemoveUsername(resolvedSession session.ResolvedSession, input *RemoveUsernameInput) (*RemoveUsernameOutput, error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID
	identityID := input.IdentityID

	var identityInfo *identity.Info
	err := s.Database.WithTx(func() (err error) {
		identityInfo, err = s.IdentityAction.RemoveIdentity(userID, identityID)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &RemoveUsernameOutput{IdentityInfo: identityInfo}, nil
}

func (s *Service) dispatchVerifyIdentityEvent(identityInfo *identity.Info, verifiedClaim *verification.Claim) error {
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
		if err := s.Events.DispatchEventOnCommit(e); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) createIdentityWithVerification(resolvedSession session.ResolvedSession, input *CreateIdentityWithVerificationInput) (*CreateIdentityWithVerificationOutput, error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID
	loginID := input.LoginID
	loginIDKey := input.LoginIDKey

	identitySpec, err := s.IdentityAction.MakeLoginIDSpec(loginIDKey, loginID)
	if err != nil {
		return nil, err
	}

	var identityInfo *identity.Info
	err = s.Database.WithTx(func() error {
		verifiedClaim, err := s.verifyIdentity(&verifyIdentityInput{
			UserID:  userID,
			Token:   input.Token,
			Channel: input.Channel,
			Code:    input.Code,
		})
		if err != nil {
			return err
		}

		// Create identity after verification
		identityInfo, _, err = s.IdentityAction.CreateIdentity(userID, identitySpec, false)
		if err != nil {
			return err
		}

		err = s.dispatchVerifyIdentityEvent(identityInfo, verifiedClaim)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &CreateIdentityWithVerificationOutput{IdentityInfo: identityInfo}, nil
}

func (s *Service) updateIdentityWithVerification(resolvedSession session.ResolvedSession, input *UpdateIdentityWithVerificationInput) (*UpdateIdentityWithVerificationOutput, error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID
	loginID := input.LoginID
	loginIDKey := input.LoginIDKey
	identityID := input.IdentityID

	var identityInfo *identity.Info
	err := s.Database.WithTx(func() error {
		verifiedClaim, err := s.verifyIdentity(&verifyIdentityInput{
			UserID:  userID,
			Token:   input.Token,
			Channel: input.Channel,
			Code:    input.Code,
		})
		if err != nil {
			return err
		}

		identitySpec, err := s.IdentityAction.MakeLoginIDSpec(loginIDKey, loginID)
		if err != nil {
			return err
		}

		identityInfo, _, err = s.IdentityAction.UpdateIdentity(userID, identityID, identitySpec, false)
		if err != nil {
			return err
		}

		err = s.dispatchVerifyIdentityEvent(identityInfo, verifiedClaim)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &UpdateIdentityWithVerificationOutput{IdentityInfo: identityInfo}, nil
}

func (s *Service) AddEmailWithVerification(resolvedSession session.ResolvedSession, input *CreateIdentityWithVerificationInput) (*CreateIdentityWithVerificationOutput, error) {
	return s.createIdentityWithVerification(resolvedSession, &CreateIdentityWithVerificationInput{
		LoginID:    input.LoginID,
		LoginIDKey: input.LoginIDKey,
		Code:       input.Code,
		Channel:    input.Channel,
		Token:      input.Token,
	})
}

func (s *Service) UpdateEmailWithVerification(resolvedSession session.ResolvedSession, input *UpdateIdentityWithVerificationInput) (*UpdateIdentityWithVerificationOutput, error) {
	return s.updateIdentityWithVerification(resolvedSession, &UpdateIdentityWithVerificationInput{
		LoginID:    input.LoginID,
		LoginIDKey: input.LoginIDKey,
		IdentityID: input.IdentityID,
		Code:       input.Code,
		Channel:    input.Channel,
		Token:      input.Token,
	})
}

func (s *Service) RemoveEmail(resolvedSession session.ResolvedSession, input *RemoveEmailInput) (*RemoveEmailOutput, error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID
	identityID := input.IdentityID

	var identityInfo *identity.Info
	err := s.Database.WithTx(func() (err error) {
		identityInfo, err = s.IdentityAction.RemoveIdentity(userID, identityID)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &RemoveEmailOutput{IdentityInfo: identityInfo}, nil
}

func (s *Service) AddPhoneNumberWithVerification(resolvedSession session.ResolvedSession, input *CreateIdentityWithVerificationInput) (*CreateIdentityWithVerificationOutput, error) {
	return s.createIdentityWithVerification(resolvedSession, &CreateIdentityWithVerificationInput{
		LoginID:    input.LoginID,
		LoginIDKey: input.LoginIDKey,
		Code:       input.Code,
		Channel:    input.Channel,
		Token:      input.Token,
	})
}

func (s *Service) UpdatePhoneNumberWithVerification(resolvedSession session.ResolvedSession, input *UpdateIdentityWithVerificationInput) (*UpdateIdentityWithVerificationOutput, error) {
	return s.updateIdentityWithVerification(resolvedSession, &UpdateIdentityWithVerificationInput{
		LoginID:    input.LoginID,
		LoginIDKey: input.LoginIDKey,
		IdentityID: input.IdentityID,
		Code:       input.Code,
		Channel:    input.Channel,
		Token:      input.Token,
	})
}

func (s *Service) RemovePhoneNumber(resolvedSession session.ResolvedSession, input *RemovePhoneNumberInput) (*RemovePhoneNumberOutput, error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID
	identityID := input.IdentityID

	var identityInfo *identity.Info
	err := s.Database.WithTx(func() (err error) {
		identityInfo, err = s.IdentityAction.RemoveIdentity(userID, identityID)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &RemovePhoneNumberOutput{IdentityInfo: identityInfo}, nil
}

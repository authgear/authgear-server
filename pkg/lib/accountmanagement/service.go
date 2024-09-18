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
	SendOTPCode(input *sendOTPCodeInput) error
	StartIdentityWithVerification(resolvedSession session.ResolvedSession, input *startIdentityWithVerificationInput) (output *StartIdentityWithVerificationOutput, err error)
	CreateIdentityWithVerification(resolvedSession session.ResolvedSession, input *CreateIdentityWithVerificationInput) (*CreateIdentityWithVerificationOutput, error)
	UpdateIdentityWithVerification(resolvedSession session.ResolvedSession, input *UpdateIdentityWithVerificationInput) (*UpdateIdentityWithVerificationOutput, error)
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
	OTPSender                 OTPSender
	OTPCodeService            OTPCodeService
	Authenticators            AuthenticatorService
	AuthenticationInfoService AuthenticationInfoService
	PasskeyService            PasskeyService
	Verification              VerificationService
	UIInfoResolver            SettingsDeleteAccountSuccessUIInfoResolver
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

func (s *Service) StartCreateEmailIdentityWithVerification(resolvedSession session.ResolvedSession, input *StartCreateIdentityWithVerificationInput) (*StartCreateIdentityWithVerificationOutput, error) {
	return s.startCreateIdentityWithVerification(resolvedSession, input)
}

func (s *Service) StartCreatePhoneNumberIdentityWithVerification(resolvedSession session.ResolvedSession, input *StartCreateIdentityWithVerificationInput) (*StartCreateIdentityWithVerificationOutput, error) {
	return s.startCreateIdentityWithVerification(resolvedSession, input)
}

func (s *Service) startCreateIdentityWithVerification(resolvedSession session.ResolvedSession, input *StartCreateIdentityWithVerificationInput) (*StartCreateIdentityWithVerificationOutput, error) {
	identitySpec, err := s.IdentityAction.MakeLoginIDSpec(input.LoginIDKey, input.LoginID)
	if err != nil {
		return nil, err
	}
	var output *StartIdentityWithVerificationOutput
	err = s.Database.WithTx(func() error {
		output, err = s.IdentityAction.StartIdentityWithVerification(resolvedSession, &startIdentityWithVerificationInput{
			LoginID:      input.LoginID,
			LoginIDKey:   input.LoginIDKey,
			IdentitySpec: identitySpec,
			Channel:      input.Channel,
			isUpdate:     false,
		})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	if !output.NeedVerification {
		return &StartCreateIdentityWithVerificationOutput{
			IdentityInfo:     output.IdentityInfo,
			NeedVerification: false,
		}, nil
	}

	var token string
	userID := resolvedSession.GetAuthenticationInfo().UserID
	loginIDType := output.IdentityInfo.LoginID.LoginIDType

	switch loginIDType {
	case model.LoginIDKeyTypeEmail:
		token, err = s.Store.GenerateToken(GenerateTokenOptions{
			UserID: userID,
			Email:  input.LoginID,
		})
	case model.LoginIDKeyTypePhone:
		token, err = s.Store.GenerateToken(GenerateTokenOptions{
			UserID:      userID,
			PhoneNumber: input.LoginID,
		})
	}
	if err != nil {
		return nil, err
	}
	return &StartCreateIdentityWithVerificationOutput{
		Token:            token,
		IdentityInfo:     output.IdentityInfo,
		NeedVerification: output.NeedVerification,
	}, nil
}

func (s *Service) StartUpdateEmailIdentityWithVerification(resolvedSession session.ResolvedSession, input *StartUpdateIdentityWithVerificationInput) (*StartUpdateIdentityWithVerificationOutput, error) {
	return s.startUpdateIdentityWithVerification(resolvedSession, input)
}

func (s *Service) StartUpdatePhoneNumberIdentityWithVerification(resolvedSession session.ResolvedSession, input *StartUpdateIdentityWithVerificationInput) (*StartUpdateIdentityWithVerificationOutput, error) {
	return s.startUpdateIdentityWithVerification(resolvedSession, input)
}

func (s *Service) startUpdateIdentityWithVerification(resolvedSession session.ResolvedSession, input *StartUpdateIdentityWithVerificationInput) (*StartUpdateIdentityWithVerificationOutput, error) {
	identitySpec, err := s.IdentityAction.MakeLoginIDSpec(input.LoginIDKey, input.LoginID)
	if err != nil {
		return nil, err
	}
	var output *StartIdentityWithVerificationOutput
	err = s.Database.WithTx(func() error {
		output, err = s.IdentityAction.StartIdentityWithVerification(resolvedSession, &startIdentityWithVerificationInput{
			LoginID:      input.LoginID,
			LoginIDKey:   input.LoginIDKey,
			IdentitySpec: identitySpec,
			Channel:      input.Channel,
			IdentityID:   input.IdentityID,
			isUpdate:     true,
		})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	if !output.NeedVerification {
		return &StartUpdateIdentityWithVerificationOutput{
			IdentityInfo:     output.IdentityInfo,
			NeedVerification: false,
		}, nil
	}

	var token string
	userID := resolvedSession.GetAuthenticationInfo().UserID
	loginIDType := output.IdentityInfo.LoginID.LoginIDType

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
		return nil, err
	}
	return &StartUpdateIdentityWithVerificationOutput{
		Token:            token,
		IdentityInfo:     output.IdentityInfo,
		NeedVerification: output.NeedVerification,
	}, nil
}

func (s *Service) ResumeCreatingEmailIdentityWithVerification(resolvedSession session.ResolvedSession, input *ResumeAddingIdentityWithVerificationInput) (*ResumeAddingIdentityWithVerificationOutput, error) {
	return s.resumeAddingIdentityWithVerification(resolvedSession, input)
}

func (s *Service) ResumeCreatingPhoneNumberIdentityWithVerification(resolvedSession session.ResolvedSession, input *ResumeAddingIdentityWithVerificationInput) (*ResumeAddingIdentityWithVerificationOutput, error) {
	return s.resumeAddingIdentityWithVerification(resolvedSession, input)
}

func (s *Service) ResumeUpdatingEmailIdentityWithVerification(resolvedSession session.ResolvedSession, input *ResumeAddingIdentityWithVerificationInput) (*ResumeAddingIdentityWithVerificationOutput, error) {
	return s.resumeAddingIdentityWithVerification(resolvedSession, input)
}

func (s *Service) ResumeUpdatingPhoneNumberIdentityWithVerification(resolvedSession session.ResolvedSession, input *ResumeAddingIdentityWithVerificationInput) (*ResumeAddingIdentityWithVerificationOutput, error) {
	return s.resumeAddingIdentityWithVerification(resolvedSession, input)
}

func (s *Service) resumeAddingIdentityWithVerification(resolvedSession session.ResolvedSession, input *ResumeAddingIdentityWithVerificationInput) (output *ResumeAddingIdentityWithVerificationOutput, err error) {
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
	identityID := token.IdentityToken.IdentityID

	switch {
	case token.IdentityToken.Email != "":
		loginID = token.IdentityToken.Email
		loginIDKeyType = model.LoginIDKeyTypeEmail
	case token.IdentityToken.PhoneNumber != "":
		loginID = token.IdentityToken.PhoneNumber
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
	err = s.IdentityAction.SendOTPCode(&sendOTPCodeInput{
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
	redirectURI := input.RedirectURI

	err := s.Database.WithTx(func() error {
		ais, err := s.Authenticators.List(
			userID,
			authenticator.KeepType(model.AuthenticatorTypePassword),
			authenticator.KeepKind(authenticator.KindPrimary),
		)
		if err != nil {
			return err
		}
		if len(ais) == 0 {
			return api.ErrNoPassword
		}
		oldInfo := ais[0]
		_, err = s.Authenticators.VerifyWithSpec(oldInfo, &authenticator.Spec{
			Password: &authenticator.PasswordSpec{
				PlainPassword: input.OldPassword,
			},
		}, nil)
		if err != nil {
			err = api.ErrInvalidCredentials
			return err
		}
		changed, newInfo, err := s.Authenticators.UpdatePassword(oldInfo, &service.UpdatePasswordOptions{
			SetPassword:    true,
			PlainPassword:  input.NewPassword,
			SetExpireAfter: true,
		})
		if err != nil {
			return err
		}
		if changed {
			err = s.Authenticators.Update(newInfo)
			if err != nil {
				return err
			}
		}
		// If is changing password with SDK.
		if input.OAuthSessionID != "" {
			authInfo := resolvedSession.GetAuthenticationInfo()
			authenticationInfoEntry := authenticationinfo.NewEntry(authInfo, input.OAuthSessionID, "")

			err = s.AuthenticationInfoService.Save(authenticationInfoEntry)
			if err != nil {
				return err
			}
			redirectURI = s.UIInfoResolver.SetAuthenticationInfoInQuery(input.RedirectURI, authenticationInfoEntry)
		}
		return nil
	})
	if err != nil {
		return nil, err
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
		identityInfo, err = s.IdentityAction.RemoveIdentity(userID, identityID)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &RemovePasskeyOutput{IdentityInfo: identityInfo}, nil
}

func (s *Service) AddIdentityBiometric(resolvedSession session.ResolvedSession, input *AddIdentityBiometricInput) (*AddIdentityBiometricOutput, error) {
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

	identitySpec := &identity.Spec{
		Type: model.IdentityTypeBiometric,
		Biometric: &identity.BiometricSpec{
			KeyID:      request.KeyID,
			Key:        string(key),
			DeviceInfo: request.DeviceInfo,
		},
	}

	var identityInfo *identity.Info
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

	return &AddIdentityBiometricOutput{IdentityInfo: identityInfo}, nil
}

func (s *Service) RemoveIdentityBiometric(resolvedSession session.ResolvedSession, input *RemoveIdentityBiometricInput) (*RemoveIdentityBiometricOuput, error) {
	identityID := input.IdentityID
	userID := resolvedSession.GetAuthenticationInfo().UserID

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

	return &RemoveIdentityBiometricOuput{IdentityInfo: identityInfo}, nil
}

func (s *Service) AddIdentityEmailWithVerification(resolvedSession session.ResolvedSession, input *AddIdentityEmailWithVerificationInput) (output *CreateIdentityWithVerificationOutput, err error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID
	defer func() {
		if err == nil {
			_, err = s.Store.ConsumeToken(input.Token)
		}
	}()

	token, err := s.Store.GetToken(input.Token)
	err = token.CheckUser(userID)
	if err != nil {
		return nil, err
	}

	identitySpec, err := s.IdentityAction.MakeLoginIDSpec(input.LoginIDKey, input.LoginID)
	if err != nil {
		return nil, err
	}

	err = s.Database.WithTx(func() (err error) {
		output, err = s.IdentityAction.CreateIdentityWithVerification(resolvedSession, &CreateIdentityWithVerificationInput{
			IdentitySpec: identitySpec,
			Code:         input.Code,
			Channel:      input.Channel,
			Token:        token,
		})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (s *Service) UpdateIdentityEmailWithVerification(resolvedSession session.ResolvedSession, input *UpdateIdentityEmailWithVerificationInput) (output *UpdateIdentityWithVerificationOutput, err error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID
	defer func() {
		if err == nil {
			_, err = s.Store.ConsumeToken(input.Token)
		}
	}()

	token, err := s.Store.GetToken(input.Token)
	err = token.CheckUser(userID)
	if err != nil {
		return nil, err
	}

	identitySpec, err := s.IdentityAction.MakeLoginIDSpec(input.LoginIDKey, input.LoginID)
	if err != nil {
		return nil, err
	}
	err = s.Database.WithTx(func() (err error) {
		output, err = s.IdentityAction.UpdateIdentityWithVerification(resolvedSession, &UpdateIdentityWithVerificationInput{
			IdentityID:   input.IdentityID,
			IdentitySpec: identitySpec,
			Code:         input.Code,
			Channel:      input.Channel,
			Token:        token,
		})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (s *Service) RemoveIdentityEmail(resolvedSession session.ResolvedSession, input *RemoveIdentityEmailInput) (*RemoveIdentityEmailOutput, error) {
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

	return &RemoveIdentityEmailOutput{IdentityInfo: identityInfo}, nil
}

func (s *Service) AddIdentityPhoneNumberWithVerification(resolvedSession session.ResolvedSession, input *AddIdentityPhoneNumberWithVerificationInput) (output *CreateIdentityWithVerificationOutput, err error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID
	defer func() {
		if err == nil {
			_, err = s.Store.ConsumeToken(input.Token)
		}
	}()

	token, err := s.Store.GetToken(input.Token)
	err = token.CheckUser(userID)
	if err != nil {
		return nil, err
	}

	identitySpec, err := s.IdentityAction.MakeLoginIDSpec(input.LoginIDKey, input.LoginID)
	if err != nil {
		return nil, err
	}

	err = s.Database.WithTx(func() (err error) {
		output, err = s.IdentityAction.CreateIdentityWithVerification(resolvedSession, &CreateIdentityWithVerificationInput{
			IdentitySpec: identitySpec,
			Code:         input.Code,
			Channel:      input.Channel,
			Token:        token,
		})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (s *Service) UpdateIdentityPhoneNumberWithVerification(resolvedSession session.ResolvedSession, input *UpdateIdentityPhoneNumberWithVerificationInput) (output *UpdateIdentityWithVerificationOutput, err error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID
	defer func() {
		if err == nil {
			_, err = s.Store.ConsumeToken(input.Token)
		}
	}()

	token, err := s.Store.GetToken(input.Token)
	err = token.CheckUser(userID)
	if err != nil {
		return nil, err
	}

	identitySpec, err := s.IdentityAction.MakeLoginIDSpec(input.LoginIDKey, input.LoginID)
	if err != nil {
		return nil, err
	}

	err = s.Database.WithTx(func() (err error) {
		output, err = s.IdentityAction.UpdateIdentityWithVerification(resolvedSession, &UpdateIdentityWithVerificationInput{
			IdentityID:   input.IdentityID,
			IdentitySpec: identitySpec,
			Code:         input.Code,
			Channel:      input.Channel,
			Token:        token,
		})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (s *Service) RemoveIdentityPhoneNumber(resolvedSession session.ResolvedSession, input *RemoveIdentityPhoneNumberInput) (*RemoveIdentityPhoneNumberOutput, error) {
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

	return &RemoveIdentityPhoneNumberOutput{IdentityInfo: identityInfo}, nil
}

func (s *Service) sendOTPCode(userID string, channel model.AuthenticatorOOBChannel, target string, isResend bool) error {
	var msgType otp.MessageType
	switch channel {
	case model.AuthenticatorOOBChannelWhatsapp:
		msgType = otp.MessageTypeWhatsappCode
	case model.AuthenticatorOOBChannelSMS:
		msgType = otp.MessageTypeVerification
	case model.AuthenticatorOOBChannelEmail:
		msgType = otp.MessageTypeVerification
	default:
		panic(fmt.Errorf("accountmanagement: unknown channel"))
	}

	msg, err := s.OTPSender.Prepare(channel, target, otp.FormCode, msgType)
	if !isResend && apierrors.IsKind(err, ratelimit.RateLimited) {
		return nil
	} else if err != nil {
		return err
	}
	defer msg.Close()

	code, err := s.OTPCodeService.GenerateOTP(
		otp.KindVerification(s.Config, channel),
		target,
		otp.FormCode,
		&otp.GenerateOptions{
			UserID: userID,
		},
	)
	// If it is not resend (switch between page), we should not send and return rate limit error to the caller.
	if !isResend && apierrors.IsKind(err, ratelimit.RateLimited) {
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

func (s *Service) verifyOTP(userID string, channel model.AuthenticatorOOBChannel, target string, code string) error {
	err := s.OTPCodeService.VerifyOTP(
		otp.KindVerification(s.Config, channel),
		target,
		code,
		&otp.VerifyOptions{
			UserID: userID,
		},
	)
	if apierrors.IsKind(err, otp.InvalidOTPCode) {
		return verification.ErrInvalidVerificationCode
	} else if err != nil {
		return err
	}
	return nil
}

func (s *Service) markClaimVerified(userID string, claimName model.ClaimName, claimValue string) error {
	verifiedClaim := s.Verification.NewVerifiedClaim(userID, string(claimName), claimValue)

	err := s.Verification.MarkClaimVerified(verifiedClaim)
	if err != nil {
		return err
	}
	return nil
}

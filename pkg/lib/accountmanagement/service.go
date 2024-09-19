package accountmanagement

import (
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
	"github.com/authgear/authgear-server/pkg/lib/translation"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
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

type UIInfoResolver interface {
	SetAuthenticationInfoInQuery(redirectURI string, e *authenticationinfo.Entry) string
}

type OTPSender interface {
	Prepare(channel model.AuthenticatorOOBChannel, target string, form otp.Form, typ translation.MessageType) (*otp.PreparedMessage, error)
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
	Events                    EventService
	OTPSender                 OTPSender
	OTPCodeService            OTPCodeService
	Authenticators            AuthenticatorService
	AuthenticationInfoService AuthenticationInfoService
	PasskeyService            PasskeyService
	Verification              VerificationService
	UIInfoResolver            UIInfoResolver
}

type StartAddingInput struct {
	UserID                                          string
	Alias                                           string
	RedirectURI                                     string
	IncludeStateAuthorizationURLAndBindStateToToken bool
}

type StartAddingOutput struct {
	Token            string `json:"token,omitempty"`
	AuthorizationURL string `json:"authorization_url,omitempty"`
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

type FinishAddingInput struct {
	UserID string
	Token  string
	Query  string
}

type FinishAddingOutput struct {
	// It is intentionally empty.
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

type ResendOTPCodeInput struct {
	UserID       string
	Channel      model.AuthenticatorOOBChannel
	LoginID      string
	isSwitchPage bool
}

func (s *Service) ResendOTPCode(input *ResendOTPCodeInput) (err error) {
	// Either it is a switch page or resend
	isResend := !input.isSwitchPage
	err = s.sendOTPCode(
		input.UserID,
		input.Channel,
		input.LoginID,
		isResend,
	)
	if err != nil {
		return err
	}

	return nil
}

type ChangePasswordInput struct {
	OAuthSessionID string
	RedirectURI    string
	OldPassword    string
	NewPassword    string
}

type ChangePasswordOutput struct {
	RedirectURI string
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

type CreateAdditionalPasswordInput struct {
	NewAuthenticatorID string
	UserID             string
	Password           string
}

func NewCreateAdditionalPasswordInput(userID string, password string) CreateAdditionalPasswordInput {
	return CreateAdditionalPasswordInput{
		NewAuthenticatorID: uuid.New(),
		UserID:             userID,
		Password:           password,
	}
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
func (s *Service) sendOTPCode(userID string, channel model.AuthenticatorOOBChannel, target string, isResend bool) error {
	var msgType translation.MessageType
	switch channel {
	case model.AuthenticatorOOBChannelWhatsapp:
		msgType = translation.MessageTypeWhatsappCode
	case model.AuthenticatorOOBChannelSMS:
		msgType = translation.MessageTypeVerification
	case model.AuthenticatorOOBChannelEmail:
		msgType = translation.MessageTypeVerification
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

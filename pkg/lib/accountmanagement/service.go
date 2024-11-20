package accountmanagement

import (
	"context"
	"fmt"

	"time"

	"github.com/authgear/oauthrelyingparty/pkg/api/oauthrelyingparty"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator/service"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/mfa"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/facade"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/translation"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type UserService interface {
	Get(ctx context.Context, id string, role accesscontrol.Role) (*model.User, error)
	UpdateMFAEnrollment(ctx context.Context, userID string, t *time.Time) error
}

type Store interface {
	GenerateToken(ctx context.Context, options GenerateTokenOptions) (string, error)
	GetToken(ctx context.Context, tokenStr string) (*Token, error)
	ConsumeToken(ctx context.Context, tokenStr string) (*Token, error)
	ConsumeToken_OAuth(ctx context.Context, tokenStr string) (*Token, error)
}

type OAuthProvider interface {
	GetProviderConfig(alias string) (oauthrelyingparty.ProviderConfig, error)

	GetAuthorizationURL(ctx context.Context, alias string, options oauthrelyingparty.GetAuthorizationURLOptions) (string, error)
	GetUserProfile(ctx context.Context, alias string, options oauthrelyingparty.GetUserProfileOptions) (oauthrelyingparty.UserProfile, error)
}

type IdentityService interface {
	New(ctx context.Context, userID string, spec *identity.Spec, options identity.NewIdentityOptions) (*identity.Info, error)
	UpdateWithSpec(ctx context.Context, is *identity.Info, spec *identity.Spec, options identity.NewIdentityOptions) (*identity.Info, error)

	Get(ctx context.Context, id string) (*identity.Info, error)
	ListByUser(ctx context.Context, userID string) ([]*identity.Info, error)
	CheckDuplicated(ctx context.Context, info *identity.Info) (dupe *identity.Info, err error)
	Create(ctx context.Context, info *identity.Info) error
	Update(ctx context.Context, oldInfo *identity.Info, newInfo *identity.Info) error
	Delete(ctx context.Context, is *identity.Info) error
}

type EventService interface {
	DispatchEventOnCommit(ctx context.Context, payload event.Payload) error
}

type AuthenticatorService interface {
	New(ctx context.Context, spec *authenticator.Spec) (*authenticator.Info, error)
	NewWithAuthenticatorID(ctx context.Context, authenticatorID string, spec *authenticator.Spec) (*authenticator.Info, error)
	UpdatePassword(ctx context.Context, authenticatorInfo *authenticator.Info, options *service.UpdatePasswordOptions) (changed bool, info *authenticator.Info, err error)

	Get(ctx context.Context, authenticatorID string) (*authenticator.Info, error)
	List(ctx context.Context, userID string, filters ...authenticator.Filter) ([]*authenticator.Info, error)
	Create(ctx context.Context, authenticatorInfo *authenticator.Info, markVerified bool) error
	Update(ctx context.Context, authenticatorInfo *authenticator.Info) error
	Delete(ctx context.Context, authenticatorInfo *authenticator.Info) error
	VerifyWithSpec(ctx context.Context, info *authenticator.Info, spec *authenticator.Spec, options *facade.VerifyOptions) (verifyResult *service.VerifyResult, err error)
}

type AuthenticationInfoService interface {
	Save(ctx context.Context, entry *authenticationinfo.Entry) error
}

type MFAService interface {
	GenerateRecoveryCodes(ctx context.Context) []string

	ReplaceRecoveryCodes(ctx context.Context, userID string, codes []string) ([]*mfa.RecoveryCode, error)
	ListRecoveryCodes(ctx context.Context, userID string) ([]*mfa.RecoveryCode, error)
}

type PasskeyService interface {
	ConsumeAttestationResponse(ctx context.Context, attestationResponse []byte) (err error)
}

type UIInfoResolver interface {
	SetAuthenticationInfoInQuery(redirectURI string, e *authenticationinfo.Entry) string
}

type OTPSender interface {
	Send(ctx context.Context, opts otp.SendOptions) error
}

type OTPCodeService interface {
	GenerateOTP(ctx context.Context, kind otp.Kind, target string, form otp.Form, opt *otp.GenerateOptions) (string, error)
	VerifyOTP(ctx context.Context, kind otp.Kind, target string, otp string, opts *otp.VerifyOptions) error
}

type VerificationService interface {
	NewVerifiedClaim(ctx context.Context, userID string, claimName string, claimValue string) *verification.Claim

	MarkClaimVerified(ctx context.Context, claim *verification.Claim) error
	GetIdentityVerificationStatus(ctx context.Context, i *identity.Info) ([]verification.ClaimStatus, error)
}

type Service struct {
	Database                  *appdb.Handle
	Config                    *config.AppConfig
	HTTPOrigin                httputil.HTTPOrigin
	Users                     UserService
	Store                     Store
	OAuthProvider             OAuthProvider
	Identities                IdentityService
	Events                    EventService
	OTPSender                 OTPSender
	OTPCodeService            OTPCodeService
	Authenticators            AuthenticatorService
	AuthenticationInfoService AuthenticationInfoService
	MFA                       MFAService
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

func (s *Service) StartAdding(ctx context.Context, input *StartAddingInput) (*StartAddingOutput, error) {
	state := ""
	if input.IncludeStateAuthorizationURLAndBindStateToToken {
		state = GenerateRandomState()
	}

	param := oauthrelyingparty.GetAuthorizationURLOptions{
		RedirectURI: input.RedirectURI,
		State:       state,
	}

	authorizationURL, err := s.OAuthProvider.GetAuthorizationURL(ctx, input.Alias, param)
	if err != nil {
		return nil, err
	}

	token, err := s.Store.GenerateToken(ctx, GenerateTokenOptions{
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

func (s *Service) FinishAdding(ctx context.Context, input *FinishAddingInput) (*FinishAddingOutput, error) {
	token, err := s.Store.ConsumeToken_OAuth(ctx, input.Token)
	if err != nil {
		return nil, err
	}

	err = token.CheckUser_OAuth(input.UserID)
	if err != nil {
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
	userProfile, err := s.OAuthProvider.GetUserProfile(ctx, token.Alias, oauthrelyingparty.GetUserProfileOptions{
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
		ctx,
		token.UserID,
		spec,
		// We are not adding Login ID here so the options is irrelevant.
		identity.NewIdentityOptions{},
	)
	if err != nil {
		return nil, err
	}

	err = s.Database.WithTx(ctx, func(ctx context.Context) error {
		_, err = s.Identities.CheckDuplicated(ctx, info)
		if err != nil {
			return err
		}

		err = s.Identities.Create(ctx, info)
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

		err = s.Events.DispatchEventOnCommit(ctx, evt)
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

func (s *Service) GetToken(ctx context.Context, resolvedSession session.ResolvedSession, tokenString string) (*Token, error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID
	token, err := s.Store.GetToken(ctx, tokenString)
	if err != nil {
		return nil, err
	}
	err = token.CheckUser(userID)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (s *Service) ResendOTPCode(ctx context.Context, resolvedSession session.ResolvedSession, tokenString string) (err error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID

	token, err := s.Store.GetToken(ctx, tokenString)
	if err != nil {
		return err
	}

	err = token.CheckUser(userID)
	if err != nil {
		return err
	}

	var target string
	var channel model.AuthenticatorOOBChannel
	if token.Identity != nil {
		if token.Identity.Email != "" {
			target = token.Identity.Email
			channel = model.AuthenticatorOOBChannelEmail
		} else if token.Identity.PhoneNumber != "" {
			target = token.Identity.PhoneNumber
			channel = model.AuthenticatorOOBChannelSMS
		}
	} else if token.Authenticator != nil {
		target = token.Authenticator.OOBOTPTarget
		channel = token.Authenticator.OOBOTPChannel
	} else {
		panic(fmt.Errorf("accountmanagement: unexpected token in resend otp code"))
	}

	err = s.Database.WithTx(ctx, func(ctx context.Context) error {
		return s.sendOTPCode(
			ctx,
			userID,
			channel,
			target,
			true,
		)
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) sendOTPCode(ctx context.Context, userID string, channel model.AuthenticatorOOBChannel, target string, isResend bool) error {
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

	code, err := s.OTPCodeService.GenerateOTP(
		ctx,
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

	err = s.OTPSender.Send(
		ctx,
		otp.SendOptions{
			Channel: channel,
			Target:  target,
			Form:    otp.FormCode,
			Type:    msgType,
			OTP:     code,
		},
	)
	if !isResend && apierrors.IsKind(err, ratelimit.RateLimited) {
		return nil
	} else if err != nil {
		return err
	}

	return nil
}

func (s *Service) VerifyOTP(ctx context.Context, userID string, channel model.AuthenticatorOOBChannel, target string, code string, skipConsume bool) error {
	err := s.OTPCodeService.VerifyOTP(
		ctx,
		otp.KindVerification(s.Config, channel),
		target,
		code,
		&otp.VerifyOptions{
			UserID:      userID,
			SkipConsume: skipConsume,
		},
	)
	if apierrors.IsKind(err, otp.InvalidOTPCode) {
		return verification.ErrInvalidVerificationCode
	} else if err != nil {
		return err
	}
	return nil
}

func (s *Service) markClaimVerified(ctx context.Context, userID string, claimName model.ClaimName, claimValue string) error {
	verifiedClaim := s.Verification.NewVerifiedClaim(ctx, userID, string(claimName), claimValue)

	err := s.Verification.MarkClaimVerified(ctx, verifiedClaim)
	if err != nil {
		return err
	}
	return nil
}

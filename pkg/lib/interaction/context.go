package interaction

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/authgear/oauthrelyingparty/pkg/api/oauthrelyingparty"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator/password"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator/service"
	"github.com/authgear/authgear-server/pkg/lib/authn/challenge"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity/anonymous"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity/biometric"
	"github.com/authgear/authgear-server/pkg/lib/authn/mfa"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/facade"
	"github.com/authgear/authgear-server/pkg/lib/feature/forgotpassword"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oauthsession"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/idpsession"
	"github.com/authgear/authgear-server/pkg/lib/webappoauth"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type IdentityService interface {
	New(ctx context.Context, userID string, spec *identity.Spec, options identity.NewIdentityOptions) (*identity.Info, error)
	UpdateWithSpec(ctx context.Context, is *identity.Info, spec *identity.Spec, options identity.NewIdentityOptions) (*identity.Info, error)

	Get(ctx context.Context, id string) (*identity.Info, error)
	SearchBySpec(ctx context.Context, spec *identity.Spec) (exactMatch *identity.Info, otherMatches []*identity.Info, err error)
	ListByUser(ctx context.Context, userID string) ([]*identity.Info, error)
	Create(ctx context.Context, is *identity.Info) error
	Update(ctx context.Context, oldInfo *identity.Info, newInfo *identity.Info) error
	Delete(ctx context.Context, is *identity.Info) error
	DeleteByAdmin(ctx context.Context, is *identity.Info) error
	CheckDuplicated(ctx context.Context, info *identity.Info) (*identity.Info, error)
}

type AuthenticatorService interface {
	New(ctx context.Context, spec *authenticator.Spec) (*authenticator.Info, error)
	NewWithAuthenticatorID(ctx context.Context, authenticatorID string, spec *authenticator.Spec) (*authenticator.Info, error)
	UpdatePassword(ctx context.Context, authenticatorInfo *authenticator.Info, options *service.UpdatePasswordOptions) (changed bool, info *authenticator.Info, err error)

	Get(ctx context.Context, id string) (*authenticator.Info, error)
	List(ctx context.Context, userID string, filters ...authenticator.Filter) ([]*authenticator.Info, error)
	Create(ctx context.Context, authenticatorInfo *authenticator.Info, markVerified bool) error
	Update(ctx context.Context, authenticatorInfo *authenticator.Info) error
	Delete(ctx context.Context, authenticatorInfo *authenticator.Info) error
	VerifyWithSpec(ctx context.Context, info *authenticator.Info, spec *authenticator.Spec, options *facade.VerifyOptions) (verifyResult *service.VerifyResult, err error)
	VerifyOneWithSpec(ctx context.Context, userID string, authenticatorType model.AuthenticatorType, infos []*authenticator.Info, spec *authenticator.Spec, options *facade.VerifyOptions) (info *authenticator.Info, verifyResult *service.VerifyResult, err error)
	ClearLockoutAttempts(ctx context.Context, userID string, usedMethods []config.AuthenticationLockoutMethod) error
	MarkOOBIdentityVerified(ctx context.Context, info *authenticator.Info) error
}

type OTPCodeService interface {
	GenerateOTP(ctx context.Context, kind otp.Kind, target string, form otp.Form, opt *otp.GenerateOptions) (string, error)
	VerifyOTP(ctx context.Context, kind otp.Kind, target string, otp string, opts *otp.VerifyOptions) error
}

type OTPSender interface {
	Send(ctx context.Context, opts otp.SendOptions) error
}

type AnonymousIdentityProvider interface {
	ParseRequestUnverified(requestJWT string) (*anonymous.Request, error)
	ParseRequest(requestJWT string, identity *identity.Anonymous) (*anonymous.Request, error)

	Get(ctx context.Context, userID string, id string) (*identity.Anonymous, error)
	GetByKeyID(ctx context.Context, keyID string) (*identity.Anonymous, error)
}

type AnonymousUserPromotionCodeStore interface {
	GetPromotionCode(ctx context.Context, codeHash string) (*anonymous.PromotionCode, error)
	DeletePromotionCode(ctx context.Context, code *anonymous.PromotionCode) error
}

type BiometricIdentityProvider interface {
	ParseRequestUnverified(requestJWT string) (*biometric.Request, error)
	ParseRequest(requestJWT string, identity *identity.Biometric) (*biometric.Request, error)

	GetByKeyID(ctx context.Context, keyID string) (*identity.Biometric, error)
}

type ChallengeProvider interface {
	Consume(ctx context.Context, token string) (*challenge.Purpose, error)
	Get(ctx context.Context, token string) (*challenge.Challenge, error)
}

type MFAService interface {
	GenerateDeviceToken(ctx context.Context) string
	CreateDeviceToken(ctx context.Context, userID string, token string) (*mfa.DeviceToken, error)
	VerifyDeviceToken(ctx context.Context, userID string, token string) error
	InvalidateAllDeviceTokens(ctx context.Context, userID string) error

	VerifyRecoveryCode(ctx context.Context, userID string, code string) (*mfa.RecoveryCode, error)
	ConsumeRecoveryCode(ctx context.Context, rc *mfa.RecoveryCode) error
	GenerateRecoveryCodes(ctx context.Context) []string
	ReplaceRecoveryCodes(ctx context.Context, userID string, codes []string) ([]*mfa.RecoveryCode, error)
	ListRecoveryCodes(ctx context.Context, userID string) ([]*mfa.RecoveryCode, error)
}

type UserService interface {
	Get(ctx context.Context, id string, role accesscontrol.Role) (*model.User, error)
	GetRaw(ctx context.Context, id string) (*user.User, error)
	Create(ctx context.Context, userID string) (*user.User, error)
	UpdateLoginTime(ctx context.Context, userID string, lastLoginAt time.Time) error
	AfterCreate(
		ctx context.Context,
		user *user.User,
		identities []*identity.Info,
		authenticators []*authenticator.Info,
		isAdminAPI bool,
	) error
}

type StdAttrsService interface {
	PopulateStandardAttributes(ctx context.Context, userID string, iden *identity.Info) error
}

type EventService interface {
	DispatchEventOnCommit(ctx context.Context, payload event.Payload) error
}

type SessionProvider interface {
	MakeSession(*session.Attrs) (*idpsession.IDPSession, string)

	Create(ctx context.Context, s *idpsession.IDPSession) error
	Reauthenticate(ctx context.Context, idpSessionID string, amr []string) error
}

type SessionManager interface {
	RevokeWithoutEvent(ctx context.Context, s session.SessionBase) error
}

type OAuthProviderFactory interface {
	GetProviderConfig(alias string) (oauthrelyingparty.ProviderConfig, error)

	GetAuthorizationURL(ctx context.Context, alias string, options oauthrelyingparty.GetAuthorizationURLOptions) (string, error)
	GetUserProfile(ctx context.Context, alias string, options oauthrelyingparty.GetUserProfileOptions) (oauthrelyingparty.UserProfile, error)
}

type OAuthRedirectURIBuilder interface {
	SSOCallbackURL(alias string) *url.URL
	SharedSSOCallbackURL() *url.URL
	WeChatAuthorizeURL(alias string) *url.URL
	WeChatCallbackEndpointURL() *url.URL
}

type OAuthStateStore interface {
	GenerateState(ctx context.Context, state *webappoauth.WebappOAuthState) (stateToken string, err error)
	PopAndRecoverState(ctx context.Context, stateToken string) (state *webappoauth.WebappOAuthState, err error)
}

type ForgotPasswordService interface {
	SendCode(ctx context.Context, loginID string, options *forgotpassword.CodeOptions) error
}

type ResetPasswordService interface {
	ResetPasswordByEndUser(ctx context.Context, code string, newPassword string) error
	ChangePasswordByAdmin(ctx context.Context, options *forgotpassword.SetPasswordOptions) error
}

type PasskeyService interface {
	ConsumeAttestationResponse(ctx context.Context, attestationResponse []byte) (err error)
	ConsumeAssertionResponse(ctx context.Context, assertionResponse []byte) (err error)
}

type VerificationService interface {
	NewVerifiedClaim(ctx context.Context, userID string, claimName string, claimValue string) *verification.Claim

	GetIdentityVerificationStatus(ctx context.Context, i *identity.Info) ([]verification.ClaimStatus, error)
	GetAuthenticatorVerificationStatus(ctx context.Context, a *authenticator.Info) (verification.AuthenticatorStatus, error)
	MarkClaimVerified(ctx context.Context, claim *verification.Claim) error
}

type CookieManager interface {
	ValueCookie(def *httputil.CookieDef, value string) *http.Cookie
	ClearCookie(def *httputil.CookieDef) *http.Cookie
}

type RateLimiter interface {
	Allow(ctx context.Context, spec ratelimit.BucketSpec) (*ratelimit.FailedReservation, error)
	Reserve(ctx context.Context, spec ratelimit.BucketSpec) (*ratelimit.Reservation, *ratelimit.FailedReservation, error)
	Cancel(ctx context.Context, r *ratelimit.Reservation)
}

type NonceService interface {
	GenerateAndSet() string
	GetAndClear() string
}

type AuthenticationInfoService interface {
	Save(ctx context.Context, entry *authenticationinfo.Entry) error
}

type OfflineGrantStore interface {
	ListClientOfflineGrants(ctx context.Context, clientID string, userID string) ([]*oauth.OfflineGrant, error)
}

type OAuthClientResolver interface {
	ResolveClient(clientID string) *config.OAuthClientConfig
}

type OAuthSessions interface {
	Get(ctx context.Context, entryID string) (*oauthsession.Entry, error)
	Save(ctx context.Context, entry *oauthsession.Entry) (err error)
}

type ContextValues struct {
	WebSessionID   string
	OAuthSessionID string
}

type Context struct {
	IsCommitting   bool   `wire:"-"`
	WebSessionID   string `wire:"-"`
	OAuthSessionID string `wire:"-"`

	Request  *http.Request
	RemoteIP httputil.RemoteIP

	Database            *appdb.SQLExecutor
	Clock               clock.Clock
	Config              *config.AppConfig
	FeatureConfig       *config.FeatureConfig
	RateLimitsEnvConfig *config.RateLimitsEnvironmentConfig
	OAuthClientResolver OAuthClientResolver

	OfflineGrants                   OfflineGrantStore
	Identities                      IdentityService
	Authenticators                  AuthenticatorService
	AnonymousIdentities             AnonymousIdentityProvider
	AnonymousUserPromotionCodeStore AnonymousUserPromotionCodeStore
	BiometricIdentities             BiometricIdentityProvider
	OTPCodeService                  OTPCodeService
	OTPSender                       OTPSender
	OAuthProviderFactory            OAuthProviderFactory
	OAuthRedirectURIBuilder         OAuthRedirectURIBuilder
	OAuthStateStore                 OAuthStateStore
	MFA                             MFAService
	ForgotPassword                  ForgotPasswordService
	ResetPassword                   ResetPasswordService
	Passkey                         PasskeyService
	Verification                    VerificationService
	RateLimiter                     RateLimiter
	PasswordGenerator               *password.Generator

	Nonces NonceService

	Challenges                ChallengeProvider
	Users                     UserService
	StdAttrsService           StdAttrsService
	Events                    EventService
	CookieManager             CookieManager
	AuthenticationInfoService AuthenticationInfoService
	Sessions                  SessionProvider
	SessionManager            SessionManager
	SessionCookie             session.CookieDef
	OAuthSessions             OAuthSessions
	MFADeviceTokenCookie      mfa.CookieDef
}

var interactionGraphSavePoint savePoint = "interaction_graph"

func (c *Context) initialize(ctx context.Context) (*Context, error) {
	interactionCtx := *c
	_, err := interactionCtx.Database.ExecWith(ctx, interactionGraphSavePoint.New())
	return &interactionCtx, err
}

func (c *Context) commit(ctx context.Context) error {
	_, err := c.Database.ExecWith(ctx, interactionGraphSavePoint.Release())
	return err
}

func (c *Context) rollback(ctx context.Context) error {
	_, err := c.Database.ExecWith(ctx, interactionGraphSavePoint.Rollback())
	return err
}

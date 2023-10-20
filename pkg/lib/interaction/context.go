package interaction

import (
	"net/http"
	"net/url"
	"time"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/challenge"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity/anonymous"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity/biometric"
	"github.com/authgear/authgear-server/pkg/lib/authn/mfa"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/authn/sso"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/facade"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/idpsession"
	"github.com/authgear/authgear-server/pkg/lib/uiparam"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type IdentityService interface {
	Get(id string) (*identity.Info, error)
	SearchBySpec(spec *identity.Spec) (exactMatch *identity.Info, otherMatches []*identity.Info, err error)
	ListByUser(userID string) ([]*identity.Info, error)
	New(userID string, spec *identity.Spec, options identity.NewIdentityOptions) (*identity.Info, error)
	UpdateWithSpec(is *identity.Info, spec *identity.Spec, options identity.NewIdentityOptions) (*identity.Info, error)
	Create(is *identity.Info) error
	Update(oldInfo *identity.Info, newInfo *identity.Info) error
	Delete(is *identity.Info) error
	CheckDuplicated(info *identity.Info) (*identity.Info, error)
}

type AuthenticatorService interface {
	Get(id string) (*authenticator.Info, error)
	List(userID string, filters ...authenticator.Filter) ([]*authenticator.Info, error)
	New(spec *authenticator.Spec) (*authenticator.Info, error)
	NewWithAuthenticatorID(authenticatorID string, spec *authenticator.Spec) (*authenticator.Info, error)
	WithSpec(authenticatorInfo *authenticator.Info, spec *authenticator.Spec) (changed bool, info *authenticator.Info, err error)
	Create(authenticatorInfo *authenticator.Info, markVerified bool) error
	Update(authenticatorInfo *authenticator.Info) error
	Delete(authenticatorInfo *authenticator.Info) error
	VerifyWithSpec(info *authenticator.Info, spec *authenticator.Spec, options *facade.VerifyOptions) (requireUpdate bool, err error)
	VerifyOneWithSpec(userID string, authenticatorType model.AuthenticatorType, infos []*authenticator.Info, spec *authenticator.Spec, options *facade.VerifyOptions) (info *authenticator.Info, requireUpdate bool, err error)
	ClearLockoutAttempts(userID string, usedMethods []config.AuthenticationLockoutMethod) error
	MarkOOBIdentityVerified(info *authenticator.Info) error
}

type OTPCodeService interface {
	GenerateOTP(kind otp.Kind, target string, form otp.Form, opt *otp.GenerateOptions) (string, error)
	VerifyOTP(kind otp.Kind, target string, otp string, opts *otp.VerifyOptions) error
}

type OTPSender interface {
	Prepare(channel model.AuthenticatorOOBChannel, target string, form otp.Form, typ otp.MessageType) (*otp.PreparedMessage, error)
	Send(msg *otp.PreparedMessage, opts otp.SendOptions) error
}

type AnonymousIdentityProvider interface {
	Get(userID string, id string) (*identity.Anonymous, error)
	ParseRequestUnverified(requestJWT string) (*anonymous.Request, error)
	GetByKeyID(keyID string) (*identity.Anonymous, error)
	ParseRequest(requestJWT string, identity *identity.Anonymous) (*anonymous.Request, error)
}

type AnonymousUserPromotionCodeStore interface {
	GetPromotionCode(codeHash string) (*anonymous.PromotionCode, error)
	DeletePromotionCode(code *anonymous.PromotionCode) error
}

type BiometricIdentityProvider interface {
	ParseRequestUnverified(requestJWT string) (*biometric.Request, error)
	GetByKeyID(keyID string) (*identity.Biometric, error)
	ParseRequest(requestJWT string, identity *identity.Biometric) (*biometric.Request, error)
}

type ChallengeProvider interface {
	Consume(token string) (*challenge.Purpose, error)
	Get(token string) (*challenge.Challenge, error)
}

type MFAService interface {
	GenerateDeviceToken() string
	CreateDeviceToken(userID string, token string) (*mfa.DeviceToken, error)
	VerifyDeviceToken(userID string, token string) error
	InvalidateAllDeviceTokens(userID string) error

	VerifyRecoveryCode(userID string, code string) (*mfa.RecoveryCode, error)
	ConsumeRecoveryCode(rc *mfa.RecoveryCode) error
	GenerateRecoveryCodes() []string
	ReplaceRecoveryCodes(userID string, codes []string) ([]*mfa.RecoveryCode, error)
	ListRecoveryCodes(userID string) ([]*mfa.RecoveryCode, error)
}

type UserService interface {
	Get(id string, role accesscontrol.Role) (*model.User, error)
	GetRaw(id string) (*user.User, error)
	Create(userID string) (*user.User, error)
	AfterCreate(
		user *user.User,
		identities []*identity.Info,
		authenticators []*authenticator.Info,
		isAdminAPI bool,
		uiParam *uiparam.T) error
	UpdateLoginTime(userID string, lastLoginAt time.Time) error
}

type StdAttrsService interface {
	PopulateStandardAttributes(userID string, iden *identity.Info) error
}

type EventService interface {
	DispatchEvent(payload event.Payload) error
}

type SessionProvider interface {
	MakeSession(*session.Attrs) (*idpsession.IDPSession, string)
	Create(*idpsession.IDPSession) error
	Reauthenticate(idpSessionID string, amr []string) error
}

type SessionManager interface {
	RevokeWithoutEvent(session.Session) error
}

type OAuthProviderFactory interface {
	NewOAuthProvider(alias string) sso.OAuthProvider
}

type OAuthRedirectURIBuilder interface {
	SSOCallbackURL(alias string) *url.URL
}

type ForgotPasswordService interface {
	SendCode(loginID string) error
}

type ResetPasswordService interface {
	ResetPassword(code string, newPassword string) error
	SetPassword(userID string, newPassword string) error
}

type PasskeyService interface {
	ConsumeAttestationResponse(attestationResponse []byte) (err error)
	ConsumeAssertionResponse(assertionResponse []byte) (err error)
}

type VerificationService interface {
	GetIdentityVerificationStatus(i *identity.Info) ([]verification.ClaimStatus, error)
	GetAuthenticatorVerificationStatus(a *authenticator.Info) (verification.AuthenticatorStatus, error)
	NewVerifiedClaim(userID string, claimName string, claimValue string) *verification.Claim
	MarkClaimVerified(claim *verification.Claim) error
}

type CookieManager interface {
	ValueCookie(def *httputil.CookieDef, value string) *http.Cookie
	ClearCookie(def *httputil.CookieDef) *http.Cookie
}

type RateLimiter interface {
	Allow(spec ratelimit.BucketSpec) error
	Reserve(spec ratelimit.BucketSpec) *ratelimit.Reservation
	Cancel(r *ratelimit.Reservation)
}

type NonceService interface {
	GenerateAndSet() string
	GetAndClear() string
}

type AuthenticationInfoService interface {
	Save(entry *authenticationinfo.Entry) error
}

type OfflineGrantStore interface {
	ListClientOfflineGrants(clientID string, userID string) ([]*oauth.OfflineGrant, error)
}

type OAuthClientResolver interface {
	ResolveClient(clientID string) *config.OAuthClientConfig
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
	MFA                             MFAService
	ForgotPassword                  ForgotPasswordService
	ResetPassword                   ResetPasswordService
	Passkey                         PasskeyService
	Verification                    VerificationService
	RateLimiter                     RateLimiter

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
	MFADeviceTokenCookie      mfa.CookieDef
}

var interactionGraphSavePoint savePoint = "interaction_graph"

func (c *Context) initialize() (*Context, error) {
	ctx := *c
	_, err := ctx.Database.ExecWith(interactionGraphSavePoint.New())
	return &ctx, err
}

func (c *Context) commit() error {
	_, err := c.Database.ExecWith(interactionGraphSavePoint.Release())
	return err
}

func (c *Context) rollback() error {
	_, err := c.Database.ExecWith(interactionGraphSavePoint.Rollback())
	return err
}

package interaction

import (
	"net/http"
	"time"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator/oob"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator/whatsapp"
	"github.com/authgear/authgear-server/pkg/lib/authn/challenge"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity/anonymous"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity/biometric"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity/loginid"
	"github.com/authgear/authgear-server/pkg/lib/authn/mfa"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/authn/sso"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/idpsession"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type IdentityService interface {
	Get(userID string, typ model.IdentityType, id string) (*identity.Info, error)
	SearchBySpec(spec *identity.Spec) (exactMatch *identity.Info, otherMatches []*identity.Info, err error)
	ListByUser(userID string) ([]*identity.Info, error)
	New(userID string, spec *identity.Spec, options identity.NewIdentityOptions) (*identity.Info, error)
	UpdateWithSpec(is *identity.Info, spec *identity.Spec, options identity.NewIdentityOptions) (*identity.Info, error)
	Create(is *identity.Info) error
	Update(info *identity.Info) error
	Delete(is *identity.Info) error
	CheckDuplicated(info *identity.Info) (*identity.Info, error)
}

type AuthenticatorService interface {
	Get(userID string, typ model.AuthenticatorType, id string) (*authenticator.Info, error)
	List(userID string, filters ...authenticator.Filter) ([]*authenticator.Info, error)
	New(spec *authenticator.Spec, secret string) (*authenticator.Info, error)
	WithSecret(authenticatorInfo *authenticator.Info, secret string) (changed bool, info *authenticator.Info, err error)
	Create(authenticatorInfo *authenticator.Info) error
	Update(authenticatorInfo *authenticator.Info) error
	Delete(authenticatorInfo *authenticator.Info) error
	VerifySecret(info *authenticator.Info, secret string) (requireUpdate bool, err error)
}

type OOBAuthenticatorProvider interface {
	GetCode(authenticatorID string) (*oob.Code, error)
	CreateCode(authenticatorID string) (*oob.Code, error)
}

type OOBCodeSender interface {
	SendCode(
		channel model.AuthenticatorOOBChannel,
		target string,
		code string,
		messageType otp.MessageType,
	) error
}

type WhatsappCodeProvider interface {
	CreateCode(phone string, appID string, webSessionID string) (*whatsapp.Code, error)
	VerifyCode(phone string, webSessionID string, consume bool) (*whatsapp.Code, error)
}

type AnonymousIdentityProvider interface {
	Get(userID string, id string) (*anonymous.Identity, error)
	ParseRequestUnverified(requestJWT string) (*anonymous.Request, error)
	GetByKeyID(keyID string) (*anonymous.Identity, error)
	ParseRequest(requestJWT string, identity *anonymous.Identity) (*anonymous.Request, error)
}

type BiometricIdentityProvider interface {
	ParseRequestUnverified(requestJWT string) (*biometric.Request, error)
	GetByKeyID(keyID string) (*biometric.Identity, error)
	ParseRequest(requestJWT string, identity *biometric.Identity) (*biometric.Request, error)
}

type ChallengeProvider interface {
	Consume(token string) (*challenge.Purpose, error)
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
	AfterCreate(user *user.User, identities []*identity.Info, authenticators []*authenticator.Info, isAdminAPI bool, webhookState string) error
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
	Delete(session.Session) error
}

type OAuthProviderFactory interface {
	NewOAuthProvider(alias string) sso.OAuthProvider
}

type ForgotPasswordService interface {
	SendCode(loginID string) error
}

type ResetPasswordService interface {
	ResetPassword(userID string, newPassword string) (oldInfo *authenticator.Info, newInfo *authenticator.Info, err error)
	ResetPasswordByCode(code string, newPassword string) (oldInfo *authenticator.Info, newInfo *authenticator.Info, err error)
	HashCode(code string) (codeHash string)
	AfterResetPasswordByCode(codeHash string) error
}

type LoginIDNormalizerFactory interface {
	NormalizerWithLoginIDType(loginIDKeyType config.LoginIDKeyType) loginid.Normalizer
}

type VerificationService interface {
	GetIdentityVerificationStatus(i *identity.Info) ([]verification.ClaimStatus, error)
	GetAuthenticatorVerificationStatus(a *authenticator.Info) (verification.AuthenticatorStatus, error)
	CreateNewCode(
		id string,
		info *identity.Info,
		webSessionID string,
		requestedByUser bool,
	) (*verification.Code, error)
	GetCode(id string) (*verification.Code, error)
	VerifyCode(id string, code string) (*verification.Code, error)
	NewVerifiedClaim(userID string, claimName string, claimValue string) *verification.Claim
	MarkClaimVerified(claim *verification.Claim) error
}

type VerificationCodeSender interface {
	SendCode(code *verification.Code) error
}

type CookieManager interface {
	ValueCookie(def *httputil.CookieDef, value string) *http.Cookie
	ClearCookie(def *httputil.CookieDef) *http.Cookie
}

type RateLimiter interface {
	TakeToken(bucket ratelimit.Bucket) error
	CheckToken(bucket ratelimit.Bucket) (pass bool, resetDuration time.Duration, err error)
}

type NonceService interface {
	GenerateAndSet() string
	GetAndClear() string
}

type AuthenticationInfoService interface {
	Save(entry *authenticationinfo.Entry) error
}

type Context struct {
	IsCommitting bool   `wire:"-"`
	WebSessionID string `wire:"-"`

	Request  *http.Request
	RemoteIP httputil.RemoteIP

	Database      *appdb.SQLExecutor
	Clock         clock.Clock
	Config        *config.AppConfig
	FeatureConfig *config.FeatureConfig

	Identities               IdentityService
	Authenticators           AuthenticatorService
	AnonymousIdentities      AnonymousIdentityProvider
	BiometricIdentities      BiometricIdentityProvider
	OOBAuthenticators        OOBAuthenticatorProvider
	OOBCodeSender            OOBCodeSender
	OAuthProviderFactory     OAuthProviderFactory
	MFA                      MFAService
	ForgotPassword           ForgotPasswordService
	ResetPassword            ResetPasswordService
	LoginIDNormalizerFactory LoginIDNormalizerFactory
	Verification             VerificationService
	VerificationCodeSender   VerificationCodeSender
	RateLimiter              RateLimiter

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
	WhatsappCodeProvider      WhatsappCodeProvider
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

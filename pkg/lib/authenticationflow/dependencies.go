package authenticationflow

import (
	"net/http"
	"time"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"
	"github.com/lestrrat-go/jwx/v2/jwt"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/accountmigration"
	"github.com/authgear/authgear-server/pkg/lib/authn/attrs"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator/service"
	"github.com/authgear/authgear-server/pkg/lib/authn/challenge"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity/anonymous"
	"github.com/authgear/authgear-server/pkg/lib/authn/mfa"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/authn/sso"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/facade"
	"github.com/authgear/authgear-server/pkg/lib/feature/forgotpassword"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/idpsession"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type IdentityService interface {
	Get(id string) (*identity.Info, error)
	SearchBySpec(spec *identity.Spec) (exactMatch *identity.Info, otherMatches []*identity.Info, err error)
	ListByClaim(name string, value string) ([]*identity.Info, error)
	ListByClaimJSONPointer(pointer jsonpointer.T, value string) ([]*identity.Info, error)
	ListByUser(userID string) ([]*identity.Info, error)
	New(userID string, spec *identity.Spec, options identity.NewIdentityOptions) (*identity.Info, error)
	UpdateWithSpec(is *identity.Info, spec *identity.Spec, options identity.NewIdentityOptions) (*identity.Info, error)
	CheckDuplicatedByUniqueKey(info *identity.Info) (*identity.Info, error)
	Create(is *identity.Info) error
	Update(oldIs *identity.Info, newIs *identity.Info) error
	Delete(is *identity.Info) error
}

type AuthenticatorService interface {
	NewWithAuthenticatorID(authenticatorID string, spec *authenticator.Spec) (*authenticator.Info, error)
	Get(authenticatorID string) (*authenticator.Info, error)
	Create(authenticatorInfo *authenticator.Info, markVerified bool) error
	Update(authenticatorInfo *authenticator.Info) error
	List(userID string, filters ...authenticator.Filter) ([]*authenticator.Info, error)
	WithSpec(authenticatorInfo *authenticator.Info, spec *authenticator.Spec) (changed bool, info *authenticator.Info, err error)
	VerifyWithSpec(info *authenticator.Info, spec *authenticator.Spec, options *facade.VerifyOptions) (verifyResult *service.VerifyResult, err error)
	VerifyOneWithSpec(userID string, authenticatorType model.AuthenticatorType, infos []*authenticator.Info, spec *authenticator.Spec, options *facade.VerifyOptions) (info *authenticator.Info, verifyResult *service.VerifyResult, err error)
	ClearLockoutAttempts(userID string, usedMethods []config.AuthenticationLockoutMethod) error
}

type OTPCodeService interface {
	GenerateOTP(kind otp.Kind, target string, form otp.Form, opt *otp.GenerateOptions) (string, error)
	VerifyOTP(kind otp.Kind, target string, otp string, opts *otp.VerifyOptions) error
	InspectState(kind otp.Kind, target string) (*otp.State, error)

	LookupCode(purpose otp.Purpose, code string) (target string, err error)
	SetSubmittedCode(kind otp.Kind, target string, code string) (*otp.State, error)
}

type OTPSender interface {
	Prepare(channel model.AuthenticatorOOBChannel, target string, form otp.Form, typ otp.MessageType) (*otp.PreparedMessage, error)
	Send(msg *otp.PreparedMessage, opts otp.SendOptions) error
}

type AnonymousIdentityService interface {
	Get(userID string, id string) (*identity.Anonymous, error)
	ParseRequestUnverified(requestJWT string) (*anonymous.Request, error)
	GetByKeyID(keyID string) (*identity.Anonymous, error)
	ParseRequest(requestJWT string, identity *identity.Anonymous) (*anonymous.Request, error)
}

type AnonymousUserPromotionCodeStore interface {
	GetPromotionCode(codeHash string) (*anonymous.PromotionCode, error)
	DeletePromotionCode(code *anonymous.PromotionCode) error
}

type VerificationService interface {
	GetClaimStatus(userID string, claimName model.ClaimName, claimValue string) (*verification.ClaimStatus, error)
	GetIdentityVerificationStatus(i *identity.Info) ([]verification.ClaimStatus, error)
	NewVerifiedClaim(userID string, claimName string, claimValue string) *verification.Claim
	MarkClaimVerified(claim *verification.Claim) error
}

type ForgotPasswordService interface {
	SendCode(loginID string, options *forgotpassword.CodeOptions) error
	IsRateLimitError(err error, target string, channel forgotpassword.CodeChannel, kind forgotpassword.CodeKind) bool
	CodeLength(target string, channel forgotpassword.CodeChannel, kind forgotpassword.CodeKind) int
	InspectState(target string, channel forgotpassword.CodeChannel, kind forgotpassword.CodeKind) (*otp.State, error)
}

type ResetPasswordService interface {
	VerifyCode(code string) (state *otp.State, err error)
	VerifyCodeWithTarget(target string, code string, channel forgotpassword.CodeChannel, kind forgotpassword.CodeKind) (state *otp.State, err error)
	ResetPassword(code string, newPassword string) error
	ResetPasswordWithTarget(target string, code string, newPassword string, channel forgotpassword.CodeChannel, kind forgotpassword.CodeKind) error
}

type RateLimiter interface {
	Allow(spec ratelimit.BucketSpec) error
	Reserve(spec ratelimit.BucketSpec) *ratelimit.Reservation
	Cancel(r *ratelimit.Reservation)
}

type EventService interface {
	DispatchEventOnCommit(payload event.Payload) error
	DispatchEventImmediately(payload event.NonBlockingPayload) error
}

type UserService interface {
	Get(id string, role accesscontrol.Role) (*model.User, error)
	GetRaw(id string) (*user.User, error)
	Create(userID string) (*user.User, error)
	UpdateLoginTime(userID string, t time.Time) error
	AfterCreate(
		user *user.User,
		identities []*identity.Info,
		authenticators []*authenticator.Info,
		isAdminAPI bool,
	) error
}

type IDPSessionService interface {
	MakeSession(*session.Attrs) (*idpsession.IDPSession, string)
	Create(*idpsession.IDPSession) error
	Reauthenticate(idpSessionID string, amr []string) error
}

type SessionService interface {
	RevokeWithoutEvent(session.Session) error
}

type StdAttrsService interface {
	PopulateStandardAttributes(userID string, iden *identity.Info) error
	UpdateStandardAttributesWithList(role accesscontrol.Role, userID string, attrs attrs.List) error
}

type CustomAttrsService interface {
	UpdateCustomAttributesWithList(role accesscontrol.Role, userID string, attrs attrs.List) error
}

type AuthenticationInfoService interface {
	Save(entry *authenticationinfo.Entry) error
}

type CookieManager interface {
	GetCookie(r *http.Request, def *httputil.CookieDef) (*http.Cookie, error)
	ValueCookie(def *httputil.CookieDef, value string) *http.Cookie
	ClearCookie(def *httputil.CookieDef) *http.Cookie
}

type AccountMigrationService interface {
	Run(migrationTokenString string) (*accountmigration.HookResponse, error)
}

type CaptchaService interface {
	VerifyToken(token string) error
}

type ChallengeService interface {
	Consume(token string) (*challenge.Purpose, error)
	Get(token string) (*challenge.Challenge, error)
}

type MFAService interface {
	GenerateRecoveryCodes() []string
	ListRecoveryCodes(userID string) ([]*mfa.RecoveryCode, error)
	ReplaceRecoveryCodes(userID string, codes []string) ([]*mfa.RecoveryCode, error)
	VerifyRecoveryCode(userID string, code string) (*mfa.RecoveryCode, error)
	ConsumeRecoveryCode(c *mfa.RecoveryCode) error

	GenerateDeviceToken() string
	CreateDeviceToken(userID string, token string) (*mfa.DeviceToken, error)
	VerifyDeviceToken(userID string, deviceToken string) error
}

type OfflineGrantStore interface {
	ListClientOfflineGrants(clientID string, userID string) ([]*oauth.OfflineGrant, error)
}

type OAuthProviderFactory interface {
	NewOAuthProvider(alias string) sso.OAuthProvider
}

type PasskeyRequestOptionsService interface {
	MakeModalRequestOptions() (*model.WebAuthnRequestOptions, error)
	MakeModalRequestOptionsWithUser(userID string) (*model.WebAuthnRequestOptions, error)
}

type PasskeyCreationOptionsService interface {
	MakeCreationOptions(userID string) (*model.WebAuthnCreationOptions, error)
}

type PasskeyService interface {
	ConsumeAttestationResponse(attestationResponse []byte) (err error)
	ConsumeAssertionResponse(assertionResponse []byte) (err error)
}

type IDTokenService interface {
	VerifyIDTokenHintWithoutClient(idTokenHint string) (jwt.Token, error)
}

type Dependencies struct {
	Config        *config.AppConfig
	FeatureConfig *config.FeatureConfig

	Clock      clock.Clock
	RemoteIP   httputil.RemoteIP
	HTTPOrigin httputil.HTTPOrigin

	HTTPRequest *http.Request

	Users                           UserService
	Identities                      IdentityService
	AnonymousIdentities             AnonymousIdentityService
	AnonymousUserPromotionCodeStore AnonymousUserPromotionCodeStore
	Authenticators                  AuthenticatorService
	MFA                             MFAService
	StdAttrsService                 StdAttrsService
	CustomAttrsService              CustomAttrsService
	OTPCodes                        OTPCodeService
	OTPSender                       OTPSender
	Verification                    VerificationService
	ForgotPassword                  ForgotPasswordService
	ResetPassword                   ResetPasswordService
	AccountMigrations               AccountMigrationService
	Challenges                      ChallengeService
	Captcha                         CaptchaService
	OAuthProviderFactory            OAuthProviderFactory
	PasskeyRequestOptionsService    PasskeyRequestOptionsService
	PasskeyCreationOptionsService   PasskeyCreationOptionsService
	PasskeyService                  PasskeyService

	IDPSessions          IDPSessionService
	Sessions             SessionService
	AuthenticationInfos  AuthenticationInfoService
	SessionCookie        session.CookieDef
	MFADeviceTokenCookie mfa.CookieDef

	Cookies CookieManager

	Events      EventService
	RateLimiter RateLimiter

	OfflineGrants OfflineGrantStore
	IDTokens      IDTokenService
}

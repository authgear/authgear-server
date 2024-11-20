package authenticationflow

import (
	"context"
	"net/http"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwt"

	"github.com/authgear/oauthrelyingparty/pkg/api/oauthrelyingparty"

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
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/facade"
	"github.com/authgear/authgear-server/pkg/lib/feature/forgotpassword"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/lib/ldap"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/idpsession"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type IdentityService interface {
	New(ctx context.Context, userID string, spec *identity.Spec, options identity.NewIdentityOptions) (*identity.Info, error)
	UpdateWithSpec(ctx context.Context, is *identity.Info, spec *identity.Spec, options identity.NewIdentityOptions) (*identity.Info, error)
	Get(ctx context.Context, id string) (*identity.Info, error)
	SearchBySpec(ctx context.Context, spec *identity.Spec) (exactMatch *identity.Info, otherMatches []*identity.Info, err error)
	ListByClaim(ctx context.Context, name string, value string) ([]*identity.Info, error)
	ListByUser(ctx context.Context, userID string) ([]*identity.Info, error)
	CheckDuplicatedByUniqueKey(ctx context.Context, info *identity.Info) (*identity.Info, error)
	Create(ctx context.Context, is *identity.Info) error
	Update(ctx context.Context, oldIs *identity.Info, newIs *identity.Info) error
	Delete(ctx context.Context, is *identity.Info) error
}

type AuthenticatorService interface {
	NewWithAuthenticatorID(ctx context.Context, authenticatorID string, spec *authenticator.Spec) (*authenticator.Info, error)
	UpdatePassword(ctx context.Context, ainfo *authenticator.Info, options *service.UpdatePasswordOptions) (changed bool, info *authenticator.Info, err error)
	Get(ctx context.Context, authenticatorID string) (*authenticator.Info, error)
	Create(ctx context.Context, authenticatorInfo *authenticator.Info, markVerified bool) error
	Update(ctx context.Context, authenticatorInfo *authenticator.Info) error
	List(ctx context.Context, userID string, filters ...authenticator.Filter) ([]*authenticator.Info, error)
	VerifyWithSpec(ctx context.Context, info *authenticator.Info, spec *authenticator.Spec, options *facade.VerifyOptions) (verifyResult *service.VerifyResult, err error)
	VerifyOneWithSpec(ctx context.Context, userID string, authenticatorType model.AuthenticatorType, infos []*authenticator.Info, spec *authenticator.Spec, options *facade.VerifyOptions) (info *authenticator.Info, verifyResult *service.VerifyResult, err error)
	ClearLockoutAttempts(ctx context.Context, userID string, usedMethods []config.AuthenticationLockoutMethod) error
}

type OTPCodeService interface {
	GenerateOTP(ctx context.Context, kind otp.Kind, target string, form otp.Form, opt *otp.GenerateOptions) (string, error)
	VerifyOTP(ctx context.Context, kind otp.Kind, target string, otp string, opts *otp.VerifyOptions) error
	InspectState(ctx context.Context, kind otp.Kind, target string) (*otp.State, error)

	LookupCode(ctx context.Context, purpose otp.Purpose, code string) (target string, err error)
	SetSubmittedCode(ctx context.Context, kind otp.Kind, target string, code string) (*otp.State, error)
}

type OTPSender interface {
	Send(ctx context.Context, opts otp.SendOptions) error
}

type AnonymousIdentityService interface {
	ParseRequestUnverified(requestJWT string) (*anonymous.Request, error)
	ParseRequest(requestJWT string, identity *identity.Anonymous) (*anonymous.Request, error)

	Get(ctx context.Context, userID string, id string) (*identity.Anonymous, error)
	GetByKeyID(ctx context.Context, keyID string) (*identity.Anonymous, error)
}

type AnonymousUserPromotionCodeStore interface {
	GetPromotionCode(ctx context.Context, codeHash string) (*anonymous.PromotionCode, error)
	DeletePromotionCode(ctx context.Context, code *anonymous.PromotionCode) error
}

type VerificationService interface {
	NewVerifiedClaim(ctx context.Context, userID string, claimName string, claimValue string) *verification.Claim
	GetClaimStatus(ctx context.Context, userID string, claimName model.ClaimName, claimValue string) (*verification.ClaimStatus, error)
	GetIdentityVerificationStatus(ctx context.Context, i *identity.Info) ([]verification.ClaimStatus, error)
	MarkClaimVerified(ctx context.Context, claim *verification.Claim) error
}

type ForgotPasswordService interface {
	IsRateLimitError(err error, target string, channel forgotpassword.CodeChannel, kind forgotpassword.CodeKind) bool
	CodeLength(target string, channel forgotpassword.CodeChannel, kind forgotpassword.CodeKind) int

	SendCode(ctx context.Context, loginID string, options *forgotpassword.CodeOptions) error
	InspectState(ctx context.Context, target string, channel forgotpassword.CodeChannel, kind forgotpassword.CodeKind) (*otp.State, error)
}

type ResetPasswordService interface {
	VerifyCode(ctx context.Context, code string) (state *otp.State, err error)
	VerifyCodeWithTarget(ctx context.Context, target string, code string, channel forgotpassword.CodeChannel, kind forgotpassword.CodeKind) (state *otp.State, err error)
	ResetPasswordByEndUser(ctx context.Context, code string, newPassword string) error
	ResetPasswordWithTarget(ctx context.Context, target string, code string, newPassword string, channel forgotpassword.CodeChannel, kind forgotpassword.CodeKind) error
}

type RateLimiter interface {
	Allow(ctx context.Context, spec ratelimit.BucketSpec) (*ratelimit.FailedReservation, error)
	Reserve(ctx context.Context, spec ratelimit.BucketSpec) (*ratelimit.Reservation, *ratelimit.FailedReservation, error)
	Cancel(ctx context.Context, r *ratelimit.Reservation)
}

type EventService interface {
	DispatchEventOnCommit(ctx context.Context, payload event.Payload) error
	DispatchEventImmediately(ctx context.Context, payload event.NonBlockingPayload) error
}

type UserService interface {
	Get(ctx context.Context, id string, role accesscontrol.Role) (*model.User, error)
	GetRaw(ctx context.Context, id string) (*user.User, error)
	Create(ctx context.Context, userID string) (*user.User, error)
	UpdateLoginTime(ctx context.Context, userID string, t time.Time) error
	UpdateMFAEnrollment(ctx context.Context, userID string, t *time.Time) error
	AfterCreate(
		ctx context.Context,
		user *user.User,
		identities []*identity.Info,
		authenticators []*authenticator.Info,
		isAdminAPI bool,
	) error
}

type IDPSessionService interface {
	MakeSession(*session.Attrs) (*idpsession.IDPSession, string)
	Create(ctx context.Context, s *idpsession.IDPSession) error
	Reauthenticate(ctx context.Context, idpSessionID string, amr []string) error
}

type SessionService interface {
	RevokeWithoutEvent(ctx context.Context, s session.SessionBase) error
}

type StdAttrsService interface {
	PopulateStandardAttributes(ctx context.Context, userID string, iden *identity.Info) error
	UpdateStandardAttributesWithList(ctx context.Context, role accesscontrol.Role, userID string, attrs attrs.List) error
}

type CustomAttrsService interface {
	UpdateCustomAttributesWithList(ctx context.Context, role accesscontrol.Role, userID string, attrs attrs.List) error
}

type AuthenticationInfoService interface {
	Save(ctx context.Context, entry *authenticationinfo.Entry) error
}

type CookieManager interface {
	GetCookie(r *http.Request, def *httputil.CookieDef) (*http.Cookie, error)
	ValueCookie(def *httputil.CookieDef, value string) *http.Cookie
	ClearCookie(def *httputil.CookieDef) *http.Cookie
}

type AccountMigrationService interface {
	Run(ctx context.Context, migrationTokenString string) (*accountmigration.HookResponse, error)
}

type BotProtectionService interface {
	Verify(ctx context.Context, response string) error
}

type CaptchaService interface {
	VerifyToken(ctx context.Context, token string) error
}

type ChallengeService interface {
	Consume(ctx context.Context, token string) (*challenge.Purpose, error)
	Get(ctx context.Context, token string) (*challenge.Challenge, error)
}

type MFAService interface {
	GenerateRecoveryCodes(ctx context.Context) []string
	ListRecoveryCodes(ctx context.Context, userID string) ([]*mfa.RecoveryCode, error)
	ReplaceRecoveryCodes(ctx context.Context, userID string, codes []string) ([]*mfa.RecoveryCode, error)
	VerifyRecoveryCode(ctx context.Context, userID string, code string) (*mfa.RecoveryCode, error)
	ConsumeRecoveryCode(ctx context.Context, c *mfa.RecoveryCode) error

	GenerateDeviceToken(ctx context.Context) string
	CreateDeviceToken(ctx context.Context, userID string, token string) (*mfa.DeviceToken, error)
	VerifyDeviceToken(ctx context.Context, userID string, deviceToken string) error
}

type OfflineGrantStore interface {
	ListClientOfflineGrants(ctx context.Context, clientID string, userID string) ([]*oauth.OfflineGrant, error)
}

type OAuthProviderFactory interface {
	GetProviderConfig(alias string) (oauthrelyingparty.ProviderConfig, error)

	GetAuthorizationURL(ctx context.Context, alias string, options oauthrelyingparty.GetAuthorizationURLOptions) (string, error)
	GetUserProfile(ctx context.Context, alias string, options oauthrelyingparty.GetUserProfileOptions) (oauthrelyingparty.UserProfile, error)
}

type PasskeyRequestOptionsService interface {
	MakeModalRequestOptions(ctx context.Context) (*model.WebAuthnRequestOptions, error)
	MakeModalRequestOptionsWithUser(ctx context.Context, userID string) (*model.WebAuthnRequestOptions, error)
}

type PasskeyCreationOptionsService interface {
	MakeCreationOptions(ctx context.Context, userID string) (*model.WebAuthnCreationOptions, error)
}

type PasskeyService interface {
	ConsumeAttestationResponse(ctx context.Context, attestationResponse []byte) (err error)
	ConsumeAssertionResponse(ctx context.Context, assertionResponse []byte) (err error)
}

type IDTokenService interface {
	VerifyIDToken(idToken string) (jwt.Token, error)
}

type LoginIDService interface {
	CheckAndNormalize(spec identity.LoginIDSpec) (normalized string, uniqueKey string, err error)
}

type LDAPService interface {
	MakeSpecFromEntry(serverConfig *config.LDAPServerConfig, loginUsername string, entry *ldap.Entry) (*identity.Spec, error)
}

type LDAPClientFactory interface {
	MakeClient(serverConfig *config.LDAPServerConfig) *ldap.Client
}

type UserFacade interface {
	GetUserIDsByLoginHint(ctx context.Context, hint *oauth.LoginHint) ([]string, error)
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
	BotProtection                   BotProtectionService
	OAuthProviderFactory            OAuthProviderFactory
	PasskeyRequestOptionsService    PasskeyRequestOptionsService
	PasskeyCreationOptionsService   PasskeyCreationOptionsService
	PasskeyService                  PasskeyService
	LoginIDs                        LoginIDService
	LDAP                            LDAPService
	LDAPClientFactory               LDAPClientFactory

	IDPSessions          IDPSessionService
	Sessions             SessionService
	AuthenticationInfos  AuthenticationInfoService
	SessionCookie        session.CookieDef
	MFADeviceTokenCookie mfa.CookieDef

	UserFacade UserFacade

	Cookies CookieManager

	Events      EventService
	RateLimiter RateLimiter

	OfflineGrants OfflineGrantStore
	IDTokens      IDTokenService
}

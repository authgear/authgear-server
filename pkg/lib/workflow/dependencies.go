package workflow

import (
	"context"
	"net/http"
	"time"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/accountmigration"
	"github.com/authgear/authgear-server/pkg/lib/authn/attrs"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator/service"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/mfa"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
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
	New(ctx context.Context, userID string, spec *identity.Spec, options identity.NewIdentityOptions) (*identity.Info, error)
	UpdateWithSpec(ctx context.Context, is *identity.Info, spec *identity.Spec, options identity.NewIdentityOptions) (*identity.Info, error)

	Get(ctx context.Context, id string) (*identity.Info, error)
	SearchBySpec(ctx context.Context, spec *identity.Spec) (exactMatch *identity.Info, otherMatches []*identity.Info, err error)
	ListByClaim(ctx context.Context, name string, value string) ([]*identity.Info, error)
	ListByUser(ctx context.Context, userID string) ([]*identity.Info, error)
	CheckDuplicated(ctx context.Context, info *identity.Info) (*identity.Info, error)
	Create(ctx context.Context, is *identity.Info) error
	Update(ctx context.Context, oldIs *identity.Info, newIs *identity.Info) error
}

type AuthenticatorService interface {
	NewWithAuthenticatorID(ctx context.Context, authenticatorID string, spec *authenticator.Spec) (*authenticator.Info, error)
	UpdatePassword(ctx context.Context, authenticatorInfo *authenticator.Info, options *service.UpdatePasswordOptions) (changed bool, info *authenticator.Info, err error)

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

type VerificationService interface {
	NewVerifiedClaim(ctx context.Context, userID string, claimName string, claimValue string) *verification.Claim
	GetIdentityVerificationStatus(ctx context.Context, i *identity.Info) ([]verification.ClaimStatus, error)
	MarkClaimVerified(ctx context.Context, claim *verification.Claim) error
}

type ForgotPasswordService interface {
	SendCode(ctx context.Context, loginID string, options *forgotpassword.CodeOptions) error
}

type ResetPasswordService interface {
	VerifyCode(ctx context.Context, code string) (state *otp.State, err error)
	ResetPasswordByEndUser(ctx context.Context, code string, newPassword string) error
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
	GetRaw(ctx context.Context, id string) (*user.User, error)
	Create(ctx context.Context, userID string) (*user.User, error)
	UpdateLoginTime(ctx context.Context, userID string, t time.Time) error
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

type EventStore interface {
	Publish(ctx context.Context, workflowID string, e Event) error
}

type AccountMigrationService interface {
	Run(ctx context.Context, migrationTokenString string) (*accountmigration.HookResponse, error)
}

type CaptchaService interface {
	VerifyToken(ctx context.Context, token string) error
}

type MFAService interface {
	GenerateRecoveryCodes(ctx context.Context) []string
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

type Dependencies struct {
	Config        *config.AppConfig
	FeatureConfig *config.FeatureConfig

	Clock    clock.Clock
	RemoteIP httputil.RemoteIP

	HTTPRequest *http.Request

	Users              UserService
	Identities         IdentityService
	Authenticators     AuthenticatorService
	MFA                MFAService
	StdAttrsService    StdAttrsService
	CustomAttrsService CustomAttrsService
	OTPCodes           OTPCodeService
	OTPSender          OTPSender
	Verification       VerificationService
	ForgotPassword     ForgotPasswordService
	ResetPassword      ResetPasswordService
	AccountMigrations  AccountMigrationService
	Captcha            CaptchaService

	IDPSessions          IDPSessionService
	Sessions             SessionService
	AuthenticationInfos  AuthenticationInfoService
	SessionCookie        session.CookieDef
	MFADeviceTokenCookie mfa.CookieDef

	Cookies CookieManager

	Events         EventService
	RateLimiter    RateLimiter
	WorkflowEvents EventStore

	OfflineGrants OfflineGrantStore
}

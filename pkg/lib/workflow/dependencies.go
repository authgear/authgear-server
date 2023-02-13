package workflow

import (
	"net/http"
	"time"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/idpsession"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type IdentityService interface {
	Get(id string) (*identity.Info, error)
	SearchBySpec(spec *identity.Spec) (exactMatch *identity.Info, otherMatches []*identity.Info, err error)
	ListByClaim(name string, value string) ([]*identity.Info, error)
	New(userID string, spec *identity.Spec, options identity.NewIdentityOptions) (*identity.Info, error)
	CheckDuplicated(info *identity.Info) (*identity.Info, error)
	Create(is *identity.Info) error
}

type AuthenticatorService interface {
	NewWithAuthenticatorID(authenticatorID string, spec *authenticator.Spec) (*authenticator.Info, error)
	Create(authenticatorInfo *authenticator.Info, markVerified bool) error
	List(userID string, filters ...authenticator.Filter) ([]*authenticator.Info, error)
	VerifyWithSpec(info *authenticator.Info, spec *authenticator.Spec) (requireUpdate bool, err error)
}

type OTPCodeService interface {
	GenerateCode(target string, otpMode otp.OTPMode, appID string, webSessionID string) (*otp.Code, error)
	VerifyCode(target string, code string) error
	VerifyMagicLinkCodeByTarget(target string, consume bool) (*otp.Code, error)
}

type OOBCodeSender interface {
	SendCode(
		channel model.AuthenticatorOOBChannel,
		target string,
		code string,
		messageType otp.MessageType,
		otpMode otp.OTPMode,
	) error
}

type VerificationService interface {
	GetIdentityVerificationStatus(i *identity.Info) ([]verification.ClaimStatus, error)
	NewVerifiedClaim(userID string, claimName string, claimValue string) *verification.Claim
	MarkClaimVerified(claim *verification.Claim) error
}

type RateLimiter interface {
	TakeToken(bucket ratelimit.Bucket) error
	CheckToken(bucket ratelimit.Bucket) (pass bool, resetDuration time.Duration, err error)
}

type EventService interface {
	DispatchEvent(payload event.Payload) error
}

type UserService interface {
	GetRaw(id string) (*user.User, error)
	Create(userID string) (*user.User, error)
	UpdateLoginTime(userID string, t time.Time) error
	AfterCreate(user *user.User, identities []*identity.Info, authenticators []*authenticator.Info, isAdminAPI bool, webhookState string) error
}

type IDPSessionService interface {
	MakeSession(*session.Attrs) (*idpsession.IDPSession, string)
	Create(*idpsession.IDPSession) error
}

type SessionService interface {
	RevokeWithoutEvent(session.Session) error
}

type StdAttrsService interface {
	PopulateStandardAttributes(userID string, iden *identity.Info) error
}

type AuthenticationInfoService interface {
	Save(entry *authenticationinfo.Entry) error
}

type CookieManager interface {
	ValueCookie(def *httputil.CookieDef, value string) *http.Cookie
	ClearCookie(def *httputil.CookieDef) *http.Cookie
}

type AntiSpamOTPCodeBucketMaker interface {
	MakeBucket(channel model.AuthenticatorOOBChannel, target string) ratelimit.Bucket
}

type Dependencies struct {
	Config        *config.AppConfig
	FeatureConfig *config.FeatureConfig

	Clock    clock.Clock
	RemoteIP httputil.RemoteIP

	Users           UserService
	Identities      IdentityService
	Authenticators  AuthenticatorService
	StdAttrsService StdAttrsService
	OTPCodes        OTPCodeService
	OOBCodeSender   OOBCodeSender
	Verification    VerificationService

	IDPSessions         IDPSessionService
	Sessions            SessionService
	AuthenticationInfos AuthenticationInfoService
	SessionCookie       session.CookieDef

	Cookies CookieManager

	Events                EventService
	RateLimiter           RateLimiter
	AntiSpamOTPCodeBucket AntiSpamOTPCodeBucketMaker
}

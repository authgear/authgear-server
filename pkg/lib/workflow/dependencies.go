package workflow

import (
	"time"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type IdentityService interface {
	Get(id string) (*identity.Info, error)
	SearchBySpec(spec *identity.Spec) (exactMatch *identity.Info, otherMatches []*identity.Info, err error)
	ListByClaim(name string, value string) ([]*identity.Info, error)
}

type OTPCodeService interface {
	GenerateCode(target string, otpMode otp.OTPMode, appID string, webSessionID string) (*otp.Code, error)
	VerifyCode(target string, code string) error
	SetUserInputtedMagicLinkCode(userInputtedCode string) (*otp.Code, error)
	VerifyMagicLinkCode(code string, consume bool) (*otp.Code, error)
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

type Dependencies struct {
	Config        *config.AppConfig
	FeatureConfig *config.FeatureConfig

	Clock    clock.Clock
	RemoteIP httputil.RemoteIP

	Identities    IdentityService
	OTPCodes      OTPCodeService
	OOBCodeSender OOBCodeSender
	Verification  VerificationService

	Events      EventService
	RateLimiter RateLimiter
}

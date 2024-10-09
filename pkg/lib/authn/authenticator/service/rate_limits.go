package service

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

const (
	VerifyPasswordPerIP        ratelimit.BucketName = "VerifyPasswordPerIP"
	VerifyPasswordPerUserPerIP ratelimit.BucketName = "VerifyPasswordPerUserPerIP"
	VerifyTOTPPerIP            ratelimit.BucketName = "VerifyTOTPPerIP"
	VerifyTOTPPerUserPerIP     ratelimit.BucketName = "VerifyTOTPPerUserPerIP"
	VerifyPasskeyPerIP         ratelimit.BucketName = "VerifyPasskeyPerIP"
	VerifyFaceRecognitionPerIP ratelimit.BucketName = "VerifyFaceRecognitionPerIP"
)

type RateLimiter interface {
	Reserve(spec ratelimit.BucketSpec) *ratelimit.Reservation
	Cancel(r *ratelimit.Reservation)
}

type Reservation struct {
	perUserPerIP *ratelimit.Reservation
	perIP        *ratelimit.Reservation
}

func (r *Reservation) Error() error {
	if r == nil {
		return nil
	}
	if err := r.perUserPerIP.Error(); err != nil {
		return err
	}
	if err := r.perIP.Error(); err != nil {
		return err
	}
	return nil
}

func (r *Reservation) Consume() {
	if r == nil {
		return
	}
	r.perUserPerIP.Consume()
	r.perIP.Consume()
}

type RateLimits struct {
	IP     httputil.RemoteIP
	Config *config.AuthenticationConfig

	RateLimiter RateLimiter
}

func (l *RateLimits) specPerIP(authType model.AuthenticatorType) ratelimit.BucketSpec {
	switch authType {
	case model.AuthenticatorTypePassword:
		config := l.Config.RateLimits.Password.PerIP
		if config.Enabled == nil {
			config = l.Config.RateLimits.General.PerIP
		}
		return ratelimit.NewBucketSpec(
			config, VerifyPasswordPerIP,
			string(l.IP),
		)

	case model.AuthenticatorTypeOOBEmail, model.AuthenticatorTypeOOBSMS:
		// OOB rate limits are handled by OTP mechanism.
		return ratelimit.BucketSpecDisabled

	case model.AuthenticatorTypeTOTP:
		config := l.Config.RateLimits.TOTP.PerIP
		if config.Enabled == nil {
			config = l.Config.RateLimits.General.PerIP
		}
		return ratelimit.NewBucketSpec(
			config, VerifyTOTPPerIP,
			string(l.IP),
		)

	case model.AuthenticatorTypePasskey:
		config := l.Config.RateLimits.Passkey.PerIP
		if config.Enabled == nil {
			config = l.Config.RateLimits.General.PerIP
		}
		return ratelimit.NewBucketSpec(
			config, VerifyPasskeyPerIP,
			string(l.IP),
		)

	case model.AuthenticatorTypeFaceRecognition:
		config := l.Config.RateLimits.General.PerIP // TODO: support face-recognition specific rate limits

		return ratelimit.NewBucketSpec(
			config, VerifyFaceRecognitionPerIP,
			string(l.IP),
		)

	default:
		panic("authenticator: unknown type: " + authType)
	}
}

func (l *RateLimits) specPerUserPerIP(userID string, authType model.AuthenticatorType) ratelimit.BucketSpec {
	switch authType {
	case model.AuthenticatorTypePassword:
		config := l.Config.RateLimits.Password.PerUserPerIP
		if config.Enabled == nil {
			config = l.Config.RateLimits.General.PerUserPerIP
		}
		return ratelimit.NewBucketSpec(
			config, VerifyPasswordPerUserPerIP,
			userID, string(l.IP),
		)

	case model.AuthenticatorTypeOOBEmail, model.AuthenticatorTypeOOBSMS:
		// OOB rate limits are handled by OTP mechanism.
		return ratelimit.BucketSpecDisabled

	case model.AuthenticatorTypeTOTP:
		config := l.Config.RateLimits.TOTP.PerUserPerIP
		if config.Enabled == nil {
			config = l.Config.RateLimits.General.PerUserPerIP
		}
		return ratelimit.NewBucketSpec(
			config, VerifyTOTPPerUserPerIP,
			userID, string(l.IP),
		)

	case model.AuthenticatorTypePasskey:
		// Per-user rate limit for passkey is handled as account enumeration rate limit,
		// since we lookup user by passkey credential ID.
		return ratelimit.BucketSpecDisabled

	case model.AuthenticatorTypeFaceRecognition:
		config := l.Config.RateLimits.General.PerUserPerIP // TODO: support face-recognition specific rate limits
		return ratelimit.NewBucketSpec(
			config, VerifyFaceRecognitionPerIP,
			userID, string(l.IP),
		)

	default:
		panic("authenticator: unknown type: " + authType)
	}
}

func (l *RateLimits) Cancel(r *Reservation) {
	l.RateLimiter.Cancel(r.perIP)
	l.RateLimiter.Cancel(r.perUserPerIP)
}

func (l *RateLimits) Reserve(userID string, authType model.AuthenticatorType) *Reservation {
	specPerUserPerIP := l.specPerUserPerIP(userID, authType)
	specPerIP := l.specPerIP(authType)

	return &Reservation{
		perUserPerIP: l.RateLimiter.Reserve(specPerUserPerIP),
		perIP:        l.RateLimiter.Reserve(specPerIP),
	}
}

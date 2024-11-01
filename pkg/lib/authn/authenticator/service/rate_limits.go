package service

import (
	"context"

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
)

type RateLimiter interface {
	Reserve(ctx context.Context, spec ratelimit.BucketSpec) (*ratelimit.Reservation, *ratelimit.FailedReservation, error)
	Cancel(ctx context.Context, r *ratelimit.Reservation)
}

type Reservation struct {
	perUserPerIP *ratelimit.Reservation
	perIP        *ratelimit.Reservation
}

func (r *Reservation) PreventCancel() {
	if r == nil {
		return
	}
	r.perUserPerIP.PreventCancel()
	r.perIP.PreventCancel()
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

	default:
		panic("authenticator: unknown type: " + authType)
	}
}

func (l *RateLimits) Cancel(ctx context.Context, r *Reservation) {
	l.RateLimiter.Cancel(ctx, r.perIP)
	l.RateLimiter.Cancel(ctx, r.perUserPerIP)
}

func (l *RateLimits) Reserve(ctx context.Context, userID string, authType model.AuthenticatorType) (*Reservation, error) {
	specPerUserPerIP := l.specPerUserPerIP(userID, authType)
	specPerIP := l.specPerIP(authType)

	rPerUserPerIP, failedPerUserPerIP, err := l.RateLimiter.Reserve(ctx, specPerUserPerIP)
	if err != nil {
		return nil, err
	}
	if err := failedPerUserPerIP.Error(); err != nil {
		return nil, err
	}

	rPerIP, failedPerIP, err := l.RateLimiter.Reserve(ctx, specPerIP)
	if err != nil {
		return nil, err
	}
	if err := failedPerIP.Error(); err != nil {
		return nil, err
	}

	return &Reservation{
		perUserPerIP: rPerUserPerIP,
		perIP:        rPerIP,
	}, nil
}

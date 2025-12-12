package service

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type RateLimiter interface {
	Reserve(ctx context.Context, spec ratelimit.BucketSpec) (*ratelimit.Reservation, *ratelimit.FailedReservation, error)
	Cancel(ctx context.Context, r *ratelimit.Reservation)
}

type Reservation struct {
	reservations []*ratelimit.Reservation
}

func (r *Reservation) PreventCancel() {
	if r == nil {
		return
	}
	for _, r := range r.reservations {
		r.PreventCancel()
	}
}

type RateLimits struct {
	IP            httputil.RemoteIP
	Config        *config.AppConfig
	FeatureConfig *config.FeatureConfig
	EnvConfig     *config.RateLimitsEnvironmentConfig

	RateLimiter RateLimiter
}

func (l *RateLimits) specs(authType model.AuthenticatorType, userID string) []*ratelimit.BucketSpec {
	opts := &ratelimit.ResolveBucketSpecOptions{
		IPAddress: string(l.IP),
		UserID:    userID,
	}

	switch authType {
	case model.AuthenticatorTypePassword:
		return ratelimit.RateLimitGroupAuthenticationPassword.ResolveBucketSpecs(l.Config, l.FeatureConfig, l.EnvConfig, opts)

	case model.AuthenticatorTypeOOBEmail, model.AuthenticatorTypeOOBSMS:
		// OOB rate limits are handled by OTP mechanism.
		return []*ratelimit.BucketSpec{}

	case model.AuthenticatorTypeTOTP:
		return ratelimit.RateLimitGroupAuthenticationTOTP.ResolveBucketSpecs(l.Config, l.FeatureConfig, l.EnvConfig, opts)

	case model.AuthenticatorTypePasskey:
		return ratelimit.RateLimitGroupAuthenticationPasskey.ResolveBucketSpecs(l.Config, l.FeatureConfig, l.EnvConfig, opts)

	default:
		panic("authenticator: unknown type: " + authType)
	}
}

func (l *RateLimits) Cancel(ctx context.Context, r *Reservation) {
	for _, resv := range r.reservations {

		l.RateLimiter.Cancel(ctx, resv)
	}
}

func (l *RateLimits) Reserve(ctx context.Context, userID string, authType model.AuthenticatorType) (*Reservation, error) {
	specs := l.specs(authType, userID)

	r := &Reservation{
		reservations: []*ratelimit.Reservation{},
	}

	for _, spec := range specs {
		spec := *spec
		resv, failedResv, err := l.RateLimiter.Reserve(ctx, spec)
		if err != nil {
			return nil, err
		}
		if err := failedResv.Error(); err != nil {
			return nil, err
		}
		r.reservations = append(r.reservations, resv)
	}

	return r, nil
}

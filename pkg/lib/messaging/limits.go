package messaging

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/lib/usage"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

const (
	usageLimitEmail    usage.LimitName = "Email"
	usageLimitSMS      usage.LimitName = "SMS"
	usageLimitWhatsapp usage.LimitName = "Whatsapp"
)

type Reservation struct {
	UsageReservation      *usage.Reservation
	RateLimitReservations []*ratelimit.Reservation
}

type UsageLimiter interface {
	Reserve(ctx context.Context, name usage.LimitName, config *config.UsageLimitConfig) (*usage.Reservation, error)
	Cancel(ctx context.Context, r *usage.Reservation)
}

type RateLimiter interface {
	Reserve(ctx context.Context, spec ratelimit.BucketSpec) (*ratelimit.Reservation, *ratelimit.FailedReservation, error)
	Cancel(ctx context.Context, r *ratelimit.Reservation)
}

type Limits struct {
	RateLimiter  RateLimiter
	UsageLimiter UsageLimiter
	RemoteIP     httputil.RemoteIP

	Config        *config.AppConfig
	FeatureConfig *config.FeatureConfig
	EnvConfig     *config.RateLimitsEnvironmentConfig
}

func (l *Limits) cancel(ctx context.Context, reservation *Reservation) {
	l.UsageLimiter.Cancel(ctx, reservation.UsageReservation)
	for _, r := range reservation.RateLimitReservations {
		l.RateLimiter.Cancel(ctx, r)
	}
}

func (l *Limits) check(
	ctx context.Context,
	reservation *Reservation,
	rl ratelimit.RateLimitGroup,
	opts *ratelimit.ResolveBucketSpecOptions,
) (err error) {
	specs := rl.ResolveBucketSpecs(l.Config, l.FeatureConfig, l.EnvConfig, opts)

	for _, spec := range specs {
		spec := *spec
		r, failed, resvErr := l.RateLimiter.Reserve(ctx, spec)
		err = resvErr
		if err != nil {
			return
		}
		if ratelimitErr := failed.Error(); ratelimitErr != nil {
			err = ratelimitErr
			return
		}
		reservation.RateLimitReservations = append(reservation.RateLimitReservations, r)
	}

	return
}

func (l *Limits) checkEmail(ctx context.Context, email string) (err error) {
	r := &Reservation{}

	defer func() {
		if err != nil {
			// Cancel any partial reservations.
			l.cancel(ctx, r)
		}
	}()

	r.UsageReservation, err = l.UsageLimiter.Reserve(ctx, usageLimitEmail, l.FeatureConfig.Messaging.EmailUsage)
	if err != nil {
		return
	}

	opts := ratelimit.ResolveBucketSpecOptions{
		IPAddress: string(l.RemoteIP),
		Target:    email,
	}

	err = l.check(ctx, r,
		ratelimit.RateLimitMessagingEmail,
		&opts,
	)
	if err != nil {
		return
	}

	return
}

func (l *Limits) checkSMS(ctx context.Context, phoneNumber string) (err error) {
	r := &Reservation{}

	defer func() {
		if err != nil {
			// Cancel any partial reservations.
			l.cancel(ctx, r)
		}
	}()

	r.UsageReservation, err = l.UsageLimiter.Reserve(ctx, usageLimitSMS, l.FeatureConfig.Messaging.SMSUsage)
	if err != nil {
		return
	}

	opts := ratelimit.ResolveBucketSpecOptions{
		IPAddress: string(l.RemoteIP),
		Target:    phoneNumber,
	}

	err = l.check(ctx, r,
		ratelimit.RateLimitMessagingSMS,
		&opts,
	)
	if err != nil {
		return
	}

	return
}

func (l *Limits) checkWhatsapp(ctx context.Context, phoneNumber string) (err error) {
	r := &Reservation{}

	defer func() {
		if err != nil {
			// Cancel any partial reservations.
			l.cancel(ctx, r)
		}
	}()

	r.UsageReservation, err = l.UsageLimiter.Reserve(ctx, usageLimitWhatsapp, l.FeatureConfig.Messaging.WhatsappUsage)
	if err != nil {
		return
	}

	opts := ratelimit.ResolveBucketSpecOptions{
		IPAddress: string(l.RemoteIP),
		Target:    phoneNumber,
	}

	// TODO: Use whatsapp specific rate limits
	err = l.check(ctx, r,
		ratelimit.RateLimitMessagingSMS,
		&opts,
	)
	if err != nil {
		return
	}

	return
}

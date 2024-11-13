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

const (
	MessagingEmailPerIP     ratelimit.BucketName = "MessagingEmailPerIP"
	MessagingEmailPerTarget ratelimit.BucketName = "MessagingEmailPerTarget"
	MessagingEmail          ratelimit.BucketName = "MessagingEmail"
	MessagingSMSPerIP       ratelimit.BucketName = "MessagingSMSPerIP"
	MessagingSMSPerTarget   ratelimit.BucketName = "MessagingSMSPerTarget"
	MessagingSMS            ratelimit.BucketName = "MessagingSMS"
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
	Logger       Logger
	RateLimiter  RateLimiter
	UsageLimiter UsageLimiter
	RemoteIP     httputil.RemoteIP

	Config        *config.MessagingRateLimitsConfig
	FeatureConfig *config.MessagingFeatureConfig
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
	global config.RateLimitsEnvironmentConfigEntry,
	feature *config.RateLimitConfig,
	local *config.RateLimitConfig,
	name ratelimit.BucketName,
	args ...string,
) (err error) {
	globalLimit := ratelimit.NewGlobalBucketSpec(global, name, args...)

	r, failed, err := l.RateLimiter.Reserve(ctx, globalLimit)
	if err != nil {
		return
	}
	if ratelimitErr := failed.Error(); ratelimitErr != nil {
		err = ratelimitErr
		return
	}
	reservation.RateLimitReservations = append(reservation.RateLimitReservations, r)

	localLimitConfig := local
	if feature.Rate() < local.Rate() {
		localLimitConfig = feature
	}
	localLimit := ratelimit.NewBucketSpec(localLimitConfig, name, args...)

	r, failed, err = l.RateLimiter.Reserve(ctx, localLimit)
	if err != nil {
		return
	}
	if ratelimitErr := failed.Error(); ratelimitErr != nil {
		err = ratelimitErr
		return
	}
	reservation.RateLimitReservations = append(reservation.RateLimitReservations, r)

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

	r.UsageReservation, err = l.UsageLimiter.Reserve(ctx, usageLimitEmail, l.FeatureConfig.EmailUsage)
	if err != nil {
		return
	}

	err = l.check(ctx, r,
		l.EnvConfig.EmailPerIP, l.FeatureConfig.RateLimits.EmailPerIP, l.Config.EmailPerIP,
		MessagingEmailPerIP, string(l.RemoteIP),
	)
	if err != nil {
		return
	}

	err = l.check(ctx, r,
		l.EnvConfig.EmailPerTarget, l.FeatureConfig.RateLimits.EmailPerTarget, l.Config.EmailPerTarget,
		MessagingEmailPerTarget, email,
	)
	if err != nil {
		return
	}

	err = l.check(ctx, r,
		l.EnvConfig.Email, l.FeatureConfig.RateLimits.Email, l.Config.Email,
		MessagingEmail,
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

	r.UsageReservation, err = l.UsageLimiter.Reserve(ctx, usageLimitSMS, l.FeatureConfig.SMSUsage)
	if err != nil {
		return
	}

	err = l.check(ctx, r,
		l.EnvConfig.SMSPerIP, l.FeatureConfig.RateLimits.SMSPerIP, l.Config.SMSPerIP,
		MessagingSMSPerIP, string(l.RemoteIP),
	)
	if err != nil {
		return
	}

	err = l.check(ctx, r,
		l.EnvConfig.SMSPerTarget, l.FeatureConfig.RateLimits.SMSPerTarget, l.Config.SMSPerTarget,
		MessagingSMSPerTarget, phoneNumber,
	)
	if err != nil {
		return
	}

	err = l.check(ctx, r,
		l.EnvConfig.SMS, l.FeatureConfig.RateLimits.SMS, l.Config.SMS,
		MessagingSMS,
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

	r.UsageReservation, err = l.UsageLimiter.Reserve(ctx, usageLimitWhatsapp, l.FeatureConfig.WhatsappUsage)
	if err != nil {
		return
	}

	// TODO: Use whatsapp specific rate limits
	err = l.check(ctx, r,
		l.EnvConfig.SMSPerIP, l.FeatureConfig.RateLimits.SMSPerIP, l.Config.SMSPerIP,
		MessagingSMSPerIP, string(l.RemoteIP),
	)
	if err != nil {
		return
	}

	// TODO: Use whatsapp specific rate limits
	err = l.check(ctx, r,
		l.EnvConfig.SMSPerTarget, l.FeatureConfig.RateLimits.SMSPerTarget, l.Config.SMSPerTarget,
		MessagingSMSPerTarget, phoneNumber,
	)
	if err != nil {
		return
	}

	// TODO: Use whatsapp specific rate limits
	err = l.check(ctx, r,
		l.EnvConfig.SMS, l.FeatureConfig.RateLimits.SMS, l.Config.SMS,
		MessagingSMS,
	)
	if err != nil {
		return
	}

	return
}

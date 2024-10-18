package messaging

import (
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

type UsageLimiter interface {
	Reserve(name usage.LimitName, config *config.UsageLimitConfig) (*usage.Reservation, error)
	Cancel(r *usage.Reservation)
}

type RateLimiter interface {
	Reserve(spec ratelimit.BucketSpec) (*ratelimit.Reservation, *ratelimit.FailedReservation, error)
	Cancel(r *ratelimit.Reservation)
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

func (l *Limits) check(
	reservations []*ratelimit.Reservation,
	global config.RateLimitsEnvironmentConfigEntry,
	feature *config.RateLimitConfig,
	local *config.RateLimitConfig,
	name ratelimit.BucketName,
	args ...string,
) (re []*ratelimit.Reservation, err error) {
	re = reservations

	globalLimit := ratelimit.NewGlobalBucketSpec(global, name, args...)

	r, failed, err := l.RateLimiter.Reserve(globalLimit)
	if err != nil {
		return
	}
	if ratelimitErr := failed.Error(); ratelimitErr != nil {
		err = ratelimitErr
		return
	}
	re = append(re, r)

	localLimitConfig := local
	if feature.Rate() < local.Rate() {
		localLimitConfig = feature
	}
	localLimit := ratelimit.NewBucketSpec(localLimitConfig, name, args...)

	r, failed, err = l.RateLimiter.Reserve(localLimit)
	if err != nil {
		return
	}
	if ratelimitErr := failed.Error(); ratelimitErr != nil {
		err = ratelimitErr
		return
	}
	re = append(re, r)

	return
}

func (l *Limits) checkEmail(email string) (msg *message, err error) {
	msg = &message{
		logger:       l.Logger,
		rateLimiter:  l.RateLimiter,
		usageLimiter: l.UsageLimiter,
	}
	defer func() {
		if err != nil {
			// Return reserved tokens
			msg.Close()
			msg = nil
		}
	}()

	msg.usageLimit, err = l.UsageLimiter.Reserve(usageLimitEmail, l.FeatureConfig.EmailUsage)
	if err != nil {
		return
	}

	msg.rateLimits, err = l.check(msg.rateLimits,
		l.EnvConfig.EmailPerIP, l.FeatureConfig.RateLimits.EmailPerIP, l.Config.EmailPerIP,
		MessagingEmailPerIP, string(l.RemoteIP),
	)
	if err != nil {
		return
	}

	msg.rateLimits, err = l.check(msg.rateLimits,
		l.EnvConfig.EmailPerTarget, l.FeatureConfig.RateLimits.EmailPerTarget, l.Config.EmailPerTarget,
		MessagingEmailPerTarget, email,
	)
	if err != nil {
		return
	}

	msg.rateLimits, err = l.check(msg.rateLimits,
		l.EnvConfig.Email, l.FeatureConfig.RateLimits.Email, l.Config.Email,
		MessagingEmail,
	)
	if err != nil {
		return
	}

	return
}

func (l *Limits) checkSMS(phoneNumber string) (msg *message, err error) {
	msg = &message{
		logger:       l.Logger,
		rateLimiter:  l.RateLimiter,
		usageLimiter: l.UsageLimiter,
	}
	defer func() {
		if err != nil {
			// Return reserved tokens
			msg.Close()
			msg = nil
		}
	}()

	msg.usageLimit, err = l.UsageLimiter.Reserve(usageLimitSMS, l.FeatureConfig.SMSUsage)
	if err != nil {
		return
	}

	msg.rateLimits, err = l.check(msg.rateLimits,
		l.EnvConfig.SMSPerIP, l.FeatureConfig.RateLimits.SMSPerIP, l.Config.SMSPerIP,
		MessagingSMSPerIP, string(l.RemoteIP),
	)
	if err != nil {
		return
	}

	msg.rateLimits, err = l.check(msg.rateLimits,
		l.EnvConfig.SMSPerTarget, l.FeatureConfig.RateLimits.SMSPerTarget, l.Config.SMSPerTarget,
		MessagingSMSPerTarget, phoneNumber,
	)
	if err != nil {
		return
	}

	msg.rateLimits, err = l.check(msg.rateLimits,
		l.EnvConfig.SMS, l.FeatureConfig.RateLimits.SMS, l.Config.SMS,
		MessagingSMS,
	)
	if err != nil {
		return
	}

	return
}

func (l *Limits) checkWhatsapp(phoneNumber string) (msg *message, err error) {
	msg = &message{
		logger:       l.Logger,
		rateLimiter:  l.RateLimiter,
		usageLimiter: l.UsageLimiter,
	}
	defer func() {
		if err != nil {
			// Return reserved tokens
			msg.Close()
			msg = nil
		}
	}()

	msg.usageLimit, err = l.UsageLimiter.Reserve(usageLimitWhatsapp, l.FeatureConfig.WhatsappUsage)
	if err != nil {
		return
	}

	// TODO: Use whatsapp specific rate limits
	msg.rateLimits, err = l.check(msg.rateLimits,
		l.EnvConfig.SMSPerIP, l.FeatureConfig.RateLimits.SMSPerIP, l.Config.SMSPerIP,
		MessagingSMSPerIP, string(l.RemoteIP),
	)
	if err != nil {
		return
	}

	// TODO: Use whatsapp specific rate limits
	msg.rateLimits, err = l.check(msg.rateLimits,
		l.EnvConfig.SMSPerTarget, l.FeatureConfig.RateLimits.SMSPerTarget, l.Config.SMSPerTarget,
		MessagingSMSPerTarget, phoneNumber,
	)
	if err != nil {
		return
	}

	// TODO: Use whatsapp specific rate limits
	msg.rateLimits, err = l.check(msg.rateLimits,
		l.EnvConfig.SMS, l.FeatureConfig.RateLimits.SMS, l.Config.SMS,
		MessagingSMS,
	)
	if err != nil {
		return
	}

	return
}

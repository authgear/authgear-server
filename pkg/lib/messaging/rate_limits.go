package messaging

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type RateLimiter interface {
	Reserve(spec ratelimit.BucketSpec) *ratelimit.Reservation
	Cancel(r *ratelimit.Reservation) error
}

// FIXME: hard usage limits
type RateLimits struct {
	Logger      Logger
	RateLimiter RateLimiter
	RemoteIP    httputil.RemoteIP

	Config        *config.MessagingRateLimitsConfig
	FeatureConfig *config.RateLimitsFeatureConfig
	EnvConfig     *config.RateLimitsEnvironmentConfig
}

func (l *RateLimits) check(
	reservations []*ratelimit.Reservation,
	global config.RateLimitsEnvironmentConfigEntry,
	feature *config.RateLimitConfig,
	local *config.RateLimitConfig,
	name string,
	args ...string,
) (re []*ratelimit.Reservation, err error) {
	re = reservations

	globalLimit := ratelimit.NewGlobalBucketSpec(global, name, args...)
	r := l.RateLimiter.Reserve(globalLimit)
	if err = r.Error(); err != nil {
		return
	}
	re = append(re, r)

	localLimitConfig := local
	if feature.Rate() < local.Rate() {
		localLimitConfig = feature
	}
	localLimit := ratelimit.NewBucketSpec(localLimitConfig, name, args...)
	r = l.RateLimiter.Reserve(localLimit)
	if err = r.Error(); err != nil {
		return
	}
	re = append(re, r)

	return
}

func (l *RateLimits) setupMessage(re []*ratelimit.Reservation, err error) (*message, error) {
	msg := &message{
		logger:       l.Logger,
		limiter:      l.RateLimiter,
		reservations: re,
	}

	if err != nil {
		// Return reserved tokens
		msg.Close()
		return nil, err
	}

	return msg, nil
}

func (l *RateLimits) checkEmail(email string) (msg *message, err error) {
	var re []*ratelimit.Reservation
	defer func() {
		msg, err = l.setupMessage(re, err)
	}()

	re, err = l.check(re,
		l.EnvConfig.EmailPerIP, l.FeatureConfig.EmailPerIP, l.Config.EmailPerIP,
		"MessagingEmailPerIP", string(l.RemoteIP),
	)
	if err != nil {
		return
	}

	re, err = l.check(re,
		l.EnvConfig.EmailPerTarget, l.FeatureConfig.EmailPerTarget, l.Config.EmailPerTarget,
		"MessagingEmailPerTarget", email,
	)
	if err != nil {
		return
	}

	re, err = l.check(re,
		l.EnvConfig.Email, l.FeatureConfig.Email, l.Config.Email,
		"MessagingEmail",
	)
	if err != nil {
		return
	}

	return
}

func (l *RateLimits) checkSMS(phoneNumber string) (msg *message, err error) {
	var re []*ratelimit.Reservation
	defer func() {
		msg, err = l.setupMessage(re, err)
	}()

	re, err = l.check(re,
		l.EnvConfig.SMSPerIP, l.FeatureConfig.SMSPerIP, l.Config.SMSPerIP,
		"MessagingSMSPerIP", string(l.RemoteIP),
	)
	if err != nil {
		return
	}

	re, err = l.check(re,
		l.EnvConfig.SMSPerTarget, l.FeatureConfig.SMSPerTarget, l.Config.SMSPerTarget,
		"MessagingSMSPerTarget", phoneNumber,
	)
	if err != nil {
		return
	}

	re, err = l.check(re,
		l.EnvConfig.SMS, l.FeatureConfig.SMS, l.Config.SMS,
		"MessagingSMS",
	)
	if err != nil {
		return
	}

	return
}

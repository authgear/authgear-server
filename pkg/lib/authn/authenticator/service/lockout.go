package service

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/lockout"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type LockoutProvider interface {
	MakeAttempt(spec lockout.BucketSpec, contributor string, attempts int) (result *lockout.MakeAttemptResult, err error)
}

func newAccountAuthenticationBucket(cfg *config.AuthenticationLockoutConfig, userID string) lockout.BucketSpec {
	isGlobal := cfg.LockoutType == config.AuthenticationLockoutTypePerUser
	return lockout.NewBucketSpec(
		lockout.BucketNameAccountAuthentication,
		cfg.MaxAttempts,
		cfg.HistoryDuration.Duration(),
		cfg.MinimumDuration.Duration(),
		cfg.MaximumDuration.Duration(),
		cfg.BackoffFactor,
		isGlobal,
		userID,
	)
}

type Lockout struct {
	Config   *config.AuthenticationLockoutConfig
	RemoteIP httputil.RemoteIP
	provider LockoutProvider
}

func (l *Lockout) Check(userID string) error {
	bucket := newAccountAuthenticationBucket(l.Config, userID)
	_, err := l.provider.MakeAttempt(bucket, string(l.RemoteIP), 0)
	if err != nil {
		return err
	}
	return nil
}

func (l *Lockout) MakeAttempt(userID string, attempts int, authenticatorType model.AuthenticatorType) error {
	switch authenticatorType {
	case model.AuthenticatorTypePassword:
		if !l.Config.Password.Enabled {
			return nil
		}
	case model.AuthenticatorTypeTOTP:
		if !l.Config.Totp.Enabled {
			return nil
		}
	case model.AuthenticatorTypeOOBEmail, model.AuthenticatorTypeOOBSMS:
		if !l.Config.OOBOTP.Enabled {
			return nil
		}
	default:
		// Not supported
		return nil
	}
	bucket := newAccountAuthenticationBucket(l.Config, userID)
	r, err := l.provider.MakeAttempt(bucket, string(l.RemoteIP), attempts)
	if err != nil {
		return err
	}
	err = r.ErrorIfLocked()
	if err != nil {
		return err
	}
	return nil
}

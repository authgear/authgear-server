package mfa

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/lockout"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type LockoutProvider interface {
	MakeAttempt(spec lockout.BucketSpec, contributor string, attempts int) (result *lockout.MakeAttemptResult, err error)
}

type Lockout struct {
	Config   *config.AuthenticationLockoutConfig
	RemoteIP httputil.RemoteIP
	provider LockoutProvider
}

func (l *Lockout) Check(userID string) error {
	bucket := lockout.NewAccountAuthenticationBucket(l.Config, userID)
	_, err := l.provider.MakeAttempt(bucket, string(l.RemoteIP), 0)
	if err != nil {
		return err
	}
	return nil
}

func (l *Lockout) MakeRecoveryCodeAttempt(userID string, attempts int) error {
	if !l.Config.RecoveryCode.Enabled {
		return nil
	}
	bucket := lockout.NewAccountAuthenticationBucket(l.Config, userID)
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

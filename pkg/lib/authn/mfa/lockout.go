package mfa

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/lockout"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type LockoutProvider interface {
	MakeAttempts(ctx context.Context, spec lockout.LockoutSpec, contributor string, attempts int) (result *lockout.MakeAttemptResult, err error)
}

type Lockout struct {
	Config   *config.AuthenticationLockoutConfig
	RemoteIP httputil.RemoteIP
	Provider LockoutProvider
}

func (l *Lockout) Check(ctx context.Context, userID string) error {
	bucket := lockout.NewAccountAuthenticationSpecForCheck(l.Config, userID)
	_, err := l.Provider.MakeAttempts(ctx, bucket, string(l.RemoteIP), 0)
	if err != nil {
		return err
	}
	return nil
}

func (l *Lockout) MakeRecoveryCodeAttempt(ctx context.Context, userID string, attempts int) error {
	bucket := lockout.NewAccountAuthenticationSpecForAttempt(l.Config, userID, []config.AuthenticationLockoutMethod{config.AuthenticationLockoutMethodRecoveryCode})
	r, err := l.Provider.MakeAttempts(ctx, bucket, string(l.RemoteIP), attempts)
	if err != nil {
		return err
	}
	err = r.ErrorIfLocked()
	if err != nil {
		return err
	}
	return nil
}

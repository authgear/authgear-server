package service

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/lockout"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type LockoutProvider interface {
	MakeAttempts(ctx context.Context, spec lockout.LockoutSpec, contributor string, attempts int) (result *lockout.MakeAttemptResult, err error)
	ClearAttempts(ctx context.Context, spec lockout.LockoutSpec, contributor string) error
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

func (l *Lockout) MakeAttempt(ctx context.Context, userID string, authenticatorType model.AuthenticatorType) error {
	method, ok := config.AuthenticationLockoutMethodFromAuthenticatorType(authenticatorType)
	if !ok {
		return nil
	}
	spec := lockout.NewAccountAuthenticationSpecForAttempt(l.Config, userID, []config.AuthenticationLockoutMethod{method})
	r, err := l.Provider.MakeAttempts(ctx, spec, string(l.RemoteIP), 1)
	if err != nil {
		return err
	}
	err = r.ErrorIfLocked()
	if err != nil {
		return err
	}
	return nil
}

func (l *Lockout) ClearAttempts(ctx context.Context, userID string, usedMethods []config.AuthenticationLockoutMethod) error {
	bucket := lockout.NewAccountAuthenticationSpecForAttempt(l.Config, userID, usedMethods)
	err := l.Provider.ClearAttempts(ctx, bucket, string(l.RemoteIP))
	if err != nil {
		return err
	}
	return nil
}

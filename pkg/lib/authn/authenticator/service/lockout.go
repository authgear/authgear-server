package service

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/lockout"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type LockoutProvider interface {
	MakeAttempts(spec lockout.BucketSpec, contributor string, attempts int) (result *lockout.MakeAttemptResult, err error)
	ClearAttempts(spec lockout.BucketSpec, contributor string) error
}

type Lockout struct {
	Config   *config.AuthenticationLockoutConfig
	RemoteIP httputil.RemoteIP
	Provider LockoutProvider
}

func (l *Lockout) Check(userID string) error {
	bucket := lockout.NewAccountAuthenticationBucket(l.Config, userID)
	_, err := l.Provider.MakeAttempts(bucket, string(l.RemoteIP), 0)
	if err != nil {
		return err
	}
	return nil
}

func (l *Lockout) checkIsParticipant(authenticatorType model.AuthenticatorType) bool {
	switch authenticatorType {
	case model.AuthenticatorTypePassword:
		if l.Config.Password.Enabled {
			return true
		}
	case model.AuthenticatorTypeTOTP:
		if l.Config.Totp.Enabled {
			return true
		}
	case model.AuthenticatorTypeOOBEmail, model.AuthenticatorTypeOOBSMS:
		if l.Config.OOBOTP.Enabled {
			return true
		}
	default:
		// Not supported
		return false
	}
	return false
}

func (l *Lockout) MakeAttempt(userID string, authenticatorType model.AuthenticatorType) error {
	if !l.checkIsParticipant(authenticatorType) {
		return nil
	}
	bucket := lockout.NewAccountAuthenticationBucket(l.Config, userID)
	r, err := l.Provider.MakeAttempts(bucket, string(l.RemoteIP), 1)
	if err != nil {
		return err
	}
	err = r.ErrorIfLocked()
	if err != nil {
		return err
	}
	return nil
}

func (l *Lockout) ClearAttempts(userID string, authenticatorTypes []model.AuthenticatorType) error {
	isParticipant := false
	for _, t := range authenticatorTypes {
		if l.checkIsParticipant(t) {
			isParticipant = true
		}
	}
	if !isParticipant {
		return nil
	}
	bucket := lockout.NewAccountAuthenticationBucket(l.Config, userID)
	err := l.Provider.ClearAttempts(bucket, string(l.RemoteIP))
	if err != nil {
		return err
	}
	return nil
}

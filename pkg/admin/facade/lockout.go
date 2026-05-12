package facade

import (
	"context"
	"sort"

	apimodel "github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	lockoutpkg "github.com/authgear/authgear-server/pkg/lib/lockout"
)

type LockoutProvider interface {
	GetStatus(ctx context.Context, spec lockoutpkg.LockoutSpec) (*lockoutpkg.LockoutStatus, error)
	ClearAll(ctx context.Context, spec lockoutpkg.LockoutSpec) error
}

type LockoutFacade struct {
	LockoutConfig *config.AuthenticationLockoutConfig
	Lockout       LockoutProvider
}

func (f *LockoutFacade) GetAccountLockoutStatus(ctx context.Context, userID string) (*apimodel.AccountLockoutStatus, error) {
	// NewAccountAuthenticationSpecForCheck returns a disabled spec when IsEnabled() is false;
	// Service.GetStatus returns {IsLocked: false} for a disabled spec without hitting Redis.
	spec := lockoutpkg.NewAccountAuthenticationSpecForCheck(f.LockoutConfig, userID)
	status, err := f.Lockout.GetStatus(ctx, spec)
	if err != nil {
		return nil, err
	}
	sort.Slice(status.LockedIPs, func(i, j int) bool {
		return status.LockedIPs[i].LockedUntil.After(status.LockedIPs[j].LockedUntil)
	})
	return &apimodel.AccountLockoutStatus{
		LockoutType: f.LockoutConfig.LockoutType,
		IsLocked:    status.IsLocked,
		LockedUntil: status.LockedUntil,
		LockedIPs:   status.LockedIPs,
	}, nil
}

func (f *LockoutFacade) ResetAccountLockout(ctx context.Context, userID string) error {
	// NewAccountAuthenticationSpecForCheck returns a disabled spec when IsEnabled() is false;
	// Service.ClearAll is a no-op for a disabled spec.
	spec := lockoutpkg.NewAccountAuthenticationSpecForCheck(f.LockoutConfig, userID)
	return f.Lockout.ClearAll(ctx, spec)
}

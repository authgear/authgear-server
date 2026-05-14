package lockout

import (
	"time"

	apimodel "github.com/authgear/authgear-server/pkg/api/model"
)

type MakeAttemptResult struct {
	spec        LockoutSpec
	LockedUntil *time.Time
}

func (m *MakeAttemptResult) ErrorIfLocked() error {
	if m.LockedUntil != nil {
		return NewErrLocked(*m.LockedUntil)
	}
	return nil
}

// LockoutStatus is the raw per-user status returned by Storage.GetStatus.
// For per_user: IsLocked and LockedUntil are populated; LockedIPs is nil.
// For per_user_per_ip: IsLocked and LockedIPs are populated; LockedUntil is nil.
type LockoutStatus struct {
	IsLocked    bool
	LockedUntil *time.Time
	LockedIPs   []apimodel.LockedIP
}

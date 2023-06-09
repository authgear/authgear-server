package lockout

import "time"

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

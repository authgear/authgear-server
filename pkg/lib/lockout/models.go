package lockout

import "time"

type MakeAttemptResult struct {
	spec        BucketSpec
	LockedUntil *time.Time
}

func (m *MakeAttemptResult) ErrorIfLocked() error {
	if m.LockedUntil != nil {
		return NewErrLocked(m.spec.Name, *m.LockedUntil)
	}
	return nil
}

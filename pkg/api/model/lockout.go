package model

import "time"

type AccountLockoutType string

const (
	AccountLockoutTypePerUser      AccountLockoutType = "per_user"
	AccountLockoutTypePerUserPerIP AccountLockoutType = "per_user_per_ip"
)

type LockedIP struct {
	IPAddress   string    `json:"ip_address"`
	LockedUntil time.Time `json:"locked_until"`
}

// AccountLockoutStatus is the admin-facing lockout state of a user.
// LockoutType is derived from config. LockedIPs is sorted by LockedUntil descending.
type AccountLockoutStatus struct {
	LockoutType AccountLockoutType `json:"lockout_type"`
	IsLocked    bool               `json:"is_locked"`
	LockedUntil *time.Time         `json:"locked_until,omitempty"` // non-nil only for per_user
	LockedIPs   []LockedIP         `json:"locked_ips"`             // non-empty only for per_user_per_ip, sorted LockedUntil desc
}

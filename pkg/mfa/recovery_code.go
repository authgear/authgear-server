package mfa

import "time"

type RecoveryCode struct {
	ID        string
	UserID    string
	Code      string
	CreatedAt time.Time
	Consumed  bool
}

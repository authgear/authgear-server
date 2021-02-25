package otp

import "time"

type CodeSendResult struct {
	Target       string
	Channel      string
	CodeLength   int
	SendCooldown int
	SentAt       time.Time
}

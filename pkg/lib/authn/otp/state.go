package otp

import "time"

type State struct {
	ExpireAt        time.Time
	CanResendAt     time.Time
	SubmittedCode   string
	TooManyAttempts bool
}

package otp

import "time"

type State struct {
	ExpireAt        time.Time
	CanResendAt     time.Time
	SubmittedCode   string
	UserID          string
	WorkflowID      string
	WebSessionID    string
	TooManyAttempts bool
}

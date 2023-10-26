package otp

import "time"

type State struct {
	ExpireAt        time.Time
	CanResendAt     time.Time
	SubmittedCode   string
	UserID          string
	TooManyAttempts bool

	WebSessionID                           string
	WorkflowID                             string
	AuthenticationFlowWebsocketChannelName string
}

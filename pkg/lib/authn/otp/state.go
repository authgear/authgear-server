package otp

import (
	"time"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"
)

type State struct {
	ExpireAt        time.Time
	CanResendAt     time.Time
	SubmittedCode   string
	UserID          string
	TooManyAttempts bool

	WebSessionID                           string
	WorkflowID                             string
	AuthenticationFlowWebsocketChannelName string
	AuthenticationFlowType                 string
	AuthenticationFlowName                 string
	AuthenticationFlowJSONPointer          jsonpointer.T
}

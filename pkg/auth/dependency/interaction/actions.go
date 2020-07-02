package interaction

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
)

// Action represents actions on an interaction that can progress the interaction state.
type Action interface {
	actionType() string
}

// ActionAuthenticate represents an attempt of authentication using the specified authenticator (e.g. password, OTP codes)
// It may also activate the authenticator if it is a pending authenticator.
type ActionAuthenticate struct {
	Authenticator authenticator.Spec `json:"spec"`
	Secret        string             `json:"secret"`
}

func (*ActionAuthenticate) actionType() string { return "authenticate" }

// ActionTriggerOOBAuthenticator represents an request to trigger the specified OOB OTP authenticator
type ActionTriggerOOBAuthenticator struct {
	Authenticator authenticator.Spec `json:"spec"`
}

func (*ActionTriggerOOBAuthenticator) actionType() string { return "trigger-oob-authenticator" }

// ActionSetupAuthenticator represents an request to setup an authenticator
type ActionSetupAuthenticator struct {
	Authenticator authenticator.Spec `json:"spec"`
	Secret        string             `json:"secret"`
}

func (*ActionSetupAuthenticator) actionType() string { return "setup-authenticator" }

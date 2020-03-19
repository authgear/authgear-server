package authn

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/session"
)

type SessionStep string

const (
	SessionStepIdentity SessionStep = "identity"
	SessionStepMFASetup SessionStep = "mfa.setup"
	SessionStepMFAAuthn SessionStep = "mfa.authn"
)

func (s SessionStep) IsMFA() bool {
	return s == SessionStepMFASetup || s == SessionStepMFAAuthn
}

// AuthnSession represents the authentication session.
// When the authentication session is finished, it converts to Session.
type Session struct {
	// The following fields are filled in step "identity"
	ClientID string `json:"client_id"`

	RequiredSteps       []SessionStep        `json:"required_steps"`
	FinishedSteps       []SessionStep        `json:"finished_steps"`
	SessionCreateReason session.CreateReason `json:"session_create_reason"`

	Attrs session.Attrs `json:"attrs"`

	AuthenticatorBearerToken string `json:"authenticator_bearer_token,omitempty"`
}

func (a *Session) SessionAttrs() *session.Attrs {
	return &a.Attrs
}

func (a *Session) IsFinished() bool {
	return len(a.RequiredSteps) == len(a.FinishedSteps)
}

func (a *Session) NextStep() (SessionStep, bool) {
	if a.IsFinished() {
		return "", false
	}
	return a.RequiredSteps[len(a.FinishedSteps)], true
}

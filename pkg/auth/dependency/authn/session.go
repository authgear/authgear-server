package authn

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/core/authn"
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
// When the authentication session is finished, it converts to AuthnSession.
// nolint: golint
type AuthnSession struct {
	// The following fields are filled in step "identity"
	ClientID   string `json:"client_id"`
	ForAuthAPI bool   `json:"for_auth_api"`

	RequiredSteps       []SessionStep            `json:"required_steps"`
	FinishedSteps       []SessionStep            `json:"finished_steps"`
	SessionCreateReason auth.SessionCreateReason `json:"session_create_reason"`

	Attrs authn.Attrs `json:"attrs"`

	AuthenticatorBearerToken string `json:"authenticator_bearer_token,omitempty"`
}

func (a *AuthnSession) AuthnAttrs() *authn.Attrs {
	return &a.Attrs
}

func (a *AuthnSession) IsFinished() bool {
	return len(a.RequiredSteps) == len(a.FinishedSteps)
}

func (a *AuthnSession) NextStep() (SessionStep, bool) {
	if a.IsFinished() {
		return "", false
	}
	return a.RequiredSteps[len(a.FinishedSteps)], true
}

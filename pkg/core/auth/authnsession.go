package auth

import (
	"fmt"
)

type AuthnSessionStep string

const (
	AuthnSessionStepIdentity AuthnSessionStep = "identity"
	AuthnSessionStepMFA      AuthnSessionStep = "mfa"
)

// AuthnSession represents the authentication session.
// When the authentication session is finished, it converts to Session.
type AuthnSession struct {
	// The following fields are filled in step "identity"
	ClientID            string             `json:"client_id"`
	UserID              string             `json:"user_id"`
	PrincipalID         string             `json:"principal_id"`
	RequiredSteps       []AuthnSessionStep `json:"required_steps"`
	FinishedSteps       []AuthnSessionStep `json:"finished_steps"`
	SessionCreateReason string             `json:"session_create_reason"`

	// The following fields are filled in step "mfa"
	AuthenticatorID          string                  `json:"authenticator_id,omitempty"`
	AuthenticatorType        AuthenticatorType       `json:"authenticator_type,omitempty"`
	AuthenticatorOOBChannel  AuthenticatorOOBChannel `json:"authenticator_oob_channel,omitempty"`
	AuthenticatorBearerToken string                  `json:"authenticator_bearer_token,omitempty"`
}

type AuthnSessionStepMFAOptions struct {
	AuthenticatorID          string
	AuthenticatorType        AuthenticatorType
	AuthenticatorOOBChannel  AuthenticatorOOBChannel
	AuthenticatorBearerToken string
}

func (a *AuthnSession) IsFinished() bool {
	return len(a.RequiredSteps) == len(a.FinishedSteps)
}

func (a *AuthnSession) NextStep() (AuthnSessionStep, bool) {
	if a.IsFinished() {
		return "", false
	}
	return a.RequiredSteps[len(a.FinishedSteps)], true
}

func (a *AuthnSession) StepMFA(opts AuthnSessionStepMFAOptions) error {
	step, ok := a.NextStep()
	if !ok || step != AuthnSessionStepMFA {
		return fmt.Errorf("expected step to be mfa")
	}
	a.AuthenticatorID = opts.AuthenticatorID
	a.AuthenticatorType = opts.AuthenticatorType
	a.AuthenticatorOOBChannel = opts.AuthenticatorOOBChannel
	a.AuthenticatorBearerToken = opts.AuthenticatorBearerToken
	a.FinishedSteps = append(a.FinishedSteps, step)
	return nil
}

func (a *AuthnSession) Session() Session {
	return Session{
		ClientID:                a.ClientID,
		UserID:                  a.UserID,
		PrincipalID:             a.PrincipalID,
		AuthenticatorID:         a.AuthenticatorID,
		AuthenticatorType:       a.AuthenticatorType,
		AuthenticatorOOBChannel: a.AuthenticatorOOBChannel,
	}
}

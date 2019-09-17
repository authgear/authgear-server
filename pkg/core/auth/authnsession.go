package auth

import (
	"time"
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
	ClientID string `json:"client_id"`
	UserID   string `json:"user_id"`

	PrincipalID        string        `json:"principal_id"`
	PrincipalType      PrincipalType `json:"principal_type"`
	PrincipalUpdatedAt time.Time     `json:"principal_updated_at"`

	RequiredSteps       []AuthnSessionStep `json:"required_steps"`
	FinishedSteps       []AuthnSessionStep `json:"finished_steps"`
	SessionCreateReason string             `json:"session_create_reason"`

	// The following fields are filled in step "mfa"
	AuthenticatorID         string                  `json:"authenticator_id,omitempty"`
	AuthenticatorType       AuthenticatorType       `json:"authenticator_type,omitempty"`
	AuthenticatorOOBChannel AuthenticatorOOBChannel `json:"authenticator_oob_channel,omitempty"`
	AuthenticatorUpdatedAt  *time.Time              `json:"authenticator_updated_at,omitempty"`

	AuthenticatorBearerToken string `json:"authenticator_bearer_token,omitempty"`
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

func (a *AuthnSession) Session() Session {
	return Session{
		ClientID:                a.ClientID,
		UserID:                  a.UserID,
		PrincipalID:             a.PrincipalID,
		PrincipalType:           a.PrincipalType,
		PrincipalUpdatedAt:      a.PrincipalUpdatedAt,
		AuthenticatorID:         a.AuthenticatorID,
		AuthenticatorType:       a.AuthenticatorType,
		AuthenticatorOOBChannel: a.AuthenticatorOOBChannel,
		AuthenticatorUpdatedAt:  a.AuthenticatorUpdatedAt,
	}
}

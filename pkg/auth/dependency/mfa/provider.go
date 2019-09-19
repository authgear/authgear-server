package mfa

import (
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
)

// Provider manipulates authenticators
type Provider interface {
	// GetRecoveryCode returns a list of recovery codes.
	GetRecoveryCode(userID string) ([]string, error)
	// GenerateRecoveryCode generates a new set of recovery codes and return it.
	GenerateRecoveryCode(userID string) ([]string, error)
	// AuthenticateRecoveryCode authenticates the user with the given code.
	AuthenticateRecoveryCode(userID string, code string) (*RecoveryCodeAuthenticator, error)

	// DeleteAllBearerToken deletes all bearer token of the given user.
	DeleteAllBearerToken(userID string) error
	// AuthenticateBearerToken authenticates the user with the given bearer token.
	AuthenticateBearerToken(userID string, token string) (*BearerTokenAuthenticator, error)

	// ListAuthenticators returns a list of authenticators.
	// Either MaskedTOTPAuthenticator or MaskedOOBAuthenticator.
	ListAuthenticators(userID string) ([]Authenticator, error)

	// CreateTOTP deletes existing inactive TOTP authenticator and creates a fresh TOTP authenticator.
	CreateTOTP(userID string, displayName string) (*TOTPAuthenticator, error)
	// ActivateTOTP activates TOTP authenticator. If this is the first authenticator,
	// a list of recovery codes are generated and returned.
	ActivateTOTP(userID string, code string) ([]string, error)
	// AuthenticateTOTP authenticates the user with the given code.
	// If generateBearerToken is true, a bearer token is generated.
	AuthenticateTOTP(userID string, code string, generateBearerToken bool) (*TOTPAuthenticator, string, error)

	// CreateOOB creates OOB authenticator.
	CreateOOB(userID string, channel coreAuth.AuthenticatorOOBChannel, phone string, email string) (*OOBAuthenticator, error)
	// TriggerOOB generate OOB Code and delivers it. The argument id is optional when
	// there is only one activated OOB authenticator.
	TriggerOOB(userID string, id string) error
	// ActivateOOB activates the OOB authenticator. If this is the first authenticator,
	// a list of recovery codes are generated and returned.
	ActivateOOB(userID string, id string, code string) ([]string, error)
	// AuthenticateOOB authenticates the user with the given code.
	// If generateBearerToken is true, a bearer token is generated.
	AuthenticateOOB(userID string, code string, generateBearerToken bool) (*OOBAuthenticator, string, error)

	// DeleteTOTP deletes authenticator.
	// It this is the last authenticator,
	// the recovery codes are also deleted.
	DeleteAuthenticator(userID string, id string) error

	// StepMFA steps forward the MFA step.
	StepMFA(authnSession *coreAuth.AuthnSession, opts coreAuth.AuthnSessionStepMFAOptions) error
}

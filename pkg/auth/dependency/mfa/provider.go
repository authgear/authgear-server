package mfa

// Provider manipulates authenticators
type Provider interface {
	// GetRecoveryCode returns a list of recovery codes.
	GetRecoveryCode(userID string) ([]string, error)
	// GenerateRecoveryCode generates a new set of recovery codes and return it.
	GenerateRecoveryCode(userID string) ([]string, error)

	// ListAuthenticators returns a list of authenticators.
	// Either MaskedTOTPAuthenticator or MaskedOOBAuthenticator.
	ListAuthenticators(userID string) ([]interface{}, error)
}

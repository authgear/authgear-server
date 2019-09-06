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

	// CreateTOTP creates TOTP authenticator.
	CreateTOTP(userID string, displayName string) (*TOTPAuthenticator, error)
	// ActivateTOTP activates TOTP authenticator. If this is the first authenticator,
	// a list of recovery codes are generated and returned.
	ActivateTOTP(userID string, id string, code string) ([]string, error)

	// DeleteTOTP deletes authenticator.
	// It this is the last authenticator,
	// the recovery codes are also deleted.
	DeleteAuthenticator(userID string, id string) error
}

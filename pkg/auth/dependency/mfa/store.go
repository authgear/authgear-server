package mfa

// Store manipulates authenticators
type Store interface {
	// GetRecoveryCode gets recovery codes.
	GetRecoveryCode(userID string) ([]RecoveryCodeAuthenticator, error)
	// GenerateRecoveryCode deletes the existing codes and generate new ones.
	GenerateRecoveryCode(userID string) ([]RecoveryCodeAuthenticator, error)

	// ListAuthenticators returns a list of authenticators.
	// Either TOTPAuthenticator or OOBAuthenticator.
	ListAuthenticators(userID string) ([]interface{}, error)
}

package mfa

// Store manipulates authenticators
type Store interface {
	GetRecoveryCode(userID string) ([]RecoveryCodeAuthenticator, error)
	// GenerateRecoveryCode deletes the existing codes and generate new ones.
	GenerateRecoveryCode(userID string) ([]RecoveryCodeAuthenticator, error)
}

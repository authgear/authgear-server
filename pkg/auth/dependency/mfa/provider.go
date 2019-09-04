package mfa

type Provider interface {
	GetRecoveryCode(userID string) ([]string, error)
	GenerateRecoveryCode(userID string) ([]string, error)
}

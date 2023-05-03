package appsecret

import "github.com/authgear/authgear-server/pkg/lib/config"

type AppSecretVisitToken struct {
	TokenID string             `json:"token_id"`
	UserID  string             `json:"user_id"`
	Secrets []config.SecretKey `json:"secrets"`
}

func NewAppSecretVisitToken(userID string, secrets []config.SecretKey) *AppSecretVisitToken {
	return &AppSecretVisitToken{
		TokenID: newSecretVisitTokenID(),
		UserID:  userID,
		Secrets: secrets,
	}
}

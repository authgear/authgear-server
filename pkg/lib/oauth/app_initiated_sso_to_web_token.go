package oauth

import (
	"time"
)

type AppInitiatedSSOToWebToken struct {
	AppID          string   `json:"app_id"`
	ClientID       string   `json:"client_id"`
	OfflineGrantID string   `json:"offline_grant_id"`
	Scopes         []string `json:"scopes"`

	CreatedAt time.Time `json:"created_at"`
	ExpireAt  time.Time `json:"expire_at"`
	TokenHash string    `json:"token_hash"`
}

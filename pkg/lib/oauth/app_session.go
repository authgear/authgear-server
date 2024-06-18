package oauth

import (
	"time"
)

type AppSession struct {
	AppID          string `json:"app_id"`
	OfflineGrantID string `json:"offline_grant_id"`

	CreatedAt        time.Time `json:"created_at"`
	ExpireAt         time.Time `json:"expire_at"`
	TokenHash        string    `json:"token_hash"`
	RefreshTokenHash string    `json:"refresh_token_hash"`
}

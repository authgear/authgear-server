package oauth

import "time"

type AccessGrant struct {
	AppID           string           `json:"app_id"`
	AuthorizationID string           `json:"authz_id"`
	SessionID       string           `json:"session_id"`
	SessionKind     GrantSessionKind `json:"session_kind"`

	CreatedAt time.Time `json:"created_at"`
	ExpireAt  time.Time `json:"expire_at"`
	Scopes    []string  `json:"scopes"`
	TokenHash string    `json:"token_hash"`
	// Only exist when session_kind is offline_grant
	RefreshTokenHash string `json:"refresh_token_hash"`
}

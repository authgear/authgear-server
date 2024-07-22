package oauth

import (
	"time"

	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
)

type SettingsActionGrant struct {
	AppID string `json:"app_id"`

	CreatedAt time.Time `json:"created_at"`
	ExpireAt  time.Time `json:"expire_at"`
	CodeHash  string    `json:"code_hash"`

	RedirectURI          string                        `json:"redirect_uri"`
	AuthorizationRequest protocol.AuthorizationRequest `json:"authorization_request"`
}

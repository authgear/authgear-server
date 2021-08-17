package oauth

import (
	"time"

	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
)

type CodeGrant struct {
	AppID              string               `json:"app_id"`
	AuthorizationID    string               `json:"authz_id"`
	IDPSessionID       string               `json:"session_id"`
	AuthenticationInfo authenticationinfo.T `json:"authentication_info"`
	IDTokenHintSID     string               `json:"id_token_hint_sid"`

	CreatedAt time.Time `json:"created_at"`
	ExpireAt  time.Time `json:"expire_at"`
	Scopes    []string  `json:"scopes"`
	CodeHash  string    `json:"code_hash"`

	RedirectURI   string `json:"redirect_uri"`
	OIDCNonce     string `json:"nonce,omitempty"`
	PKCEChallenge string `json:"challenge,omitempty"`
}

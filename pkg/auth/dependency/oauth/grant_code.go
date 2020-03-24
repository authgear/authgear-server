package oauth

import "time"

type CodeGrant struct {
	AppID           string `json:"app_id"`
	AuthorizationID string `json:"authz_id"`
	SessionID       string `json:"session_id"`

	CreatedAt time.Time `json:"created_at"`
	ExpireAt  time.Time `json:"expire_at"`
	Scopes    []string  `json:"scopes"`
	CodeHash  string    `json:"code_hash"`

	RedirectURI   string `json:"redirect_uri"`
	OIDCNonce     string `json:"nonce,omitempty"`
	PKCEChallenge string `json:"challenge,omitempty"`
}

var _ Grant = &CodeGrant{}

func (g *CodeGrant) Session() (kind GrantSessionKind, id string) {
	return GrantSessionKindSession, g.SessionID
}

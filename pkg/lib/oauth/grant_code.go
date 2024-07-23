package oauth

import (
	"crypto/subtle"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/dpop"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
)

type CodeGrant struct {
	AppID              string               `json:"app_id"`
	AuthorizationID    string               `json:"authz_id"`
	AuthenticationInfo authenticationinfo.T `json:"authentication_info"`
	IDTokenHintSID     string               `json:"id_token_hint_sid"`

	CreatedAt time.Time `json:"created_at"`
	ExpireAt  time.Time `json:"expire_at"`
	CodeHash  string    `json:"code_hash"`
	DPoPJKT   string    `json:"dpop_jkt"`

	RedirectURI          string                        `json:"redirect_uri"`
	AuthorizationRequest protocol.AuthorizationRequest `json:"authorization_request"`
}

func (g *CodeGrant) MatchDPoPJKT(proof *dpop.DPoPProof) bool {
	if g.DPoPJKT == "" {
		// Not binded, always ok
		return true
	}
	if proof == nil {
		return false
	}
	if subtle.ConstantTimeCompare([]byte(proof.JKT), []byte(g.DPoPJKT)) == 1 {
		return true
	}
	return false
}

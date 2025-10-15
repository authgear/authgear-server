package oauth

import (
	"crypto/subtle"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
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

	// IdentitySpecs is for supporting include_identity_attributes_in_id_token.
	IdentitySpecs []*identity.Spec `json:"identity_specs,omitzero"`
}

func (g *CodeGrant) MatchDPoPJKT(proof *dpop.DPoPProof) *dpop.UnmatchedJKTError {
	if g.DPoPJKT == "" {
		// Not binded, always ok
		return nil
	}
	if proof == nil {
		return dpop.NewErrUnmatchedJKT("expect DPoP proof exist to use the code grant",
			&g.DPoPJKT,
			nil,
		)
	}
	if subtle.ConstantTimeCompare([]byte(proof.JKT), []byte(g.DPoPJKT)) == 1 {
		return nil
	}
	return dpop.NewErrUnmatchedJKT("failed to match DPoP JKT of code grant",
		&g.DPoPJKT,
		&proof.JKT,
	)
}

package oauth

import (
	"time"

	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/lib/session"
)

type CodeGrant struct {
	AppID              string               `json:"app_id"`
	AuthorizationID    string               `json:"authz_id"`
	SessionType        session.Type         `json:"session_type"`
	SessionID          string               `json:"session_id"`
	AuthenticationInfo authenticationinfo.T `json:"authentication_info"`
	IDTokenHintSID     string               `json:"id_token_hint_sid"`

	CreatedAt time.Time `json:"created_at"`
	ExpireAt  time.Time `json:"expire_at"`
	CodeHash  string    `json:"code_hash"`

	RedirectURI          string                        `json:"redirect_uri"`
	AuthorizationRequest protocol.AuthorizationRequest `json:"authorization_request"`
}

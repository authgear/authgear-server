package sso

import (
	"github.com/skygeario/skygear-server/pkg/auth/model"
)

// LoginState stores login specific state.
type LoginState struct {
	MergeRealm      string                `json:"merge_realm,omitempty"`
	OnUserDuplicate model.OnUserDuplicate `json:"on_user_duplicate,omitempty"`
}

// LinkState stores link specific state.
type LinkState struct {
	UserID string `json:"user_id,omitempty"`
}

// OAuthAuthorizationCodeFlowState stores OAuth Authorization Code flow state.
type OAuthAuthorizationCodeFlowState struct {
	UXMode      UXMode `json:"ux_mode,omitempty"`
	CallbackURL string `json:"callback_url,omitempty"`
}

// State is an opaque value used by the client to maintain
// state between the request and callback.
// See https://tools.ietf.org/html/rfc6749#section-4.1.1
type State struct {
	LoginState
	LinkState
	OAuthAuthorizationCodeFlowState
	Action string `json:"action,omitempty"`
	// CodeChallenge is borrowed from PKCE.
	// See https://www.oauth.com/oauth2-servers/pkce/authorization-request/
	CodeChallenge string `json:"code_challenge"`
	Nonce         string `json:"nonce"`
	APIClientID   string `json:"api_client_id"`
}

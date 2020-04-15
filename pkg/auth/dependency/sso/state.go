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

// State is an opaque value used by the client to maintain
// state between the request and callback.
// See https://tools.ietf.org/html/rfc6749#section-4.1.1
type State struct {
	LoginState
	LinkState
	Extra       map[string]string `json:"extra,omitempty"`
	Action      string            `json:"action,omitempty"`
	UXMode      UXMode            `json:"ux_mode,omitempty"`
	HashedNonce string            `json:"hashed_nonce"`
	APIClientID string            `json:"api_client_id"`
}

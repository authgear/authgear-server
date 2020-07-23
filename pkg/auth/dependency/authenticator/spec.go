package authenticator

import "github.com/authgear/authgear-server/pkg/core/authn"

type Spec struct {
	UserID string                  `json:"user_id"`
	Type   authn.AuthenticatorType `json:"type"`
	Props  map[string]interface{}  `json:"props"`
}

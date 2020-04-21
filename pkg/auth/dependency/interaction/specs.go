package interaction

import "github.com/skygeario/skygear-server/pkg/core/authn"

type IdentitySpec struct {
	ID     string                 `json:"-"`
	Type   authn.IdentityType     `json:"type"`
	Claims map[string]interface{} `json:"claims"`
}

type AuthenticatorSpec struct {
	ID    string                 `json:"-"`
	Type  AuthenticatorType      `json:"type"`
	Props map[string]interface{} `json:"props"`
}

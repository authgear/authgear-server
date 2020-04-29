package interaction

import "github.com/skygeario/skygear-server/pkg/core/authn"

type IdentityRef struct {
	ID   string             `json:"id"`
	Type authn.IdentityType `json:"type"`
}

type IdentitySpec struct {
	Type   authn.IdentityType     `json:"type"`
	Claims map[string]interface{} `json:"claims"`
}

type AuthenticatorRef struct {
	ID   string                  `json:"id"`
	Type authn.AuthenticatorType `json:"type"`
}

type AuthenticatorSpec struct {
	Type  authn.AuthenticatorType `json:"type"`
	Props map[string]interface{}  `json:"props"`
}

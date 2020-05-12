package authenticator

import "github.com/skygeario/skygear-server/pkg/core/authn"

type Spec struct {
	Type  authn.AuthenticatorType `json:"type"`
	Props map[string]interface{}  `json:"props"`
}

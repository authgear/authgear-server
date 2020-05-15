package identity

import "github.com/skygeario/skygear-server/pkg/core/authn"

type Spec struct {
	Type   authn.IdentityType     `json:"type"`
	Claims map[string]interface{} `json:"claims"`
}

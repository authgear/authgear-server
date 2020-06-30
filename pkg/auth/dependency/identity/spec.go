package identity

import "github.com/authgear/authgear-server/pkg/core/authn"

type Spec struct {
	Type   authn.IdentityType     `json:"type"`
	Claims map[string]interface{} `json:"claims"`
}

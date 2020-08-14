package identity

import (
	"github.com/authgear/authgear-server/pkg/lib/authn"
)

type Spec struct {
	Type   authn.IdentityType     `json:"type"`
	Claims map[string]interface{} `json:"claims"`
}

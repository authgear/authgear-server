package authenticator

import (
	"github.com/authgear/authgear-server/pkg/lib/authn"
)

type Spec struct {
	UserID    string                  `json:"user_id"`
	Type      authn.AuthenticatorType `json:"type"`
	IsDefault bool                    `json:"is_default"`
	Kind      Kind                    `json:"kind"`
	Claims    map[string]interface{}  `json:"claims"`
}

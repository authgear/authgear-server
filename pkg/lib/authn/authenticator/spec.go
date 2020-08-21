package authenticator

import (
	"github.com/authgear/authgear-server/pkg/lib/authn"
)

type Spec struct {
	UserID string                  `json:"user_id"`
	Type   authn.AuthenticatorType `json:"type"`
	Tag    []string                `json:"tag,omitempty"`
	Claims map[string]interface{}  `json:"claims"`
}

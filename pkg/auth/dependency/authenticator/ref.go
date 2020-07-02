package authenticator

import "github.com/authgear/authgear-server/pkg/core/authn"

type Ref struct {
	ID   string                  `json:"id"`
	Type authn.AuthenticatorType `json:"type"`
}

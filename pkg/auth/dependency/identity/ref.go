package identity

import "github.com/skygeario/skygear-server/pkg/core/authn"

type Ref struct {
	ID   string             `json:"id"`
	Type authn.IdentityType `json:"type"`
}

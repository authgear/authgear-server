package authenticator

import (
	"github.com/authgear/authgear-server/pkg/lib/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn"
)

type Ref struct {
	model.Meta
	UserID string
	Type   authn.AuthenticatorType
}

func (r *Ref) ToRef() *Ref { return r }

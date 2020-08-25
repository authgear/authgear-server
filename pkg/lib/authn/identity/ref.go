package identity

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn"
)

type Ref struct {
	model.Meta
	UserID string
	Type   authn.IdentityType
}

func (r *Ref) ToRef() *Ref { return r }

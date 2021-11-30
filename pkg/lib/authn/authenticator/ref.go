package authenticator

import (
	"github.com/authgear/authgear-server/pkg/api/model"
)

type Ref struct {
	model.Meta
	UserID string
	Type   model.AuthenticatorType
}

func (r *Ref) ToRef() *Ref { return r }

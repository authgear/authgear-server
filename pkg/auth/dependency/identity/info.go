package identity

import "github.com/skygeario/skygear-server/pkg/core/authn"

type Info struct {
	ID       string                 `json:"id"`
	Type     authn.IdentityType     `json:"type"`
	Claims   map[string]interface{} `json:"claims"`
	Identity interface{}            `json:"-"`
}

func (i *Info) ToSpec() Spec {
	return Spec{Type: i.Type, Claims: i.Claims}
}

func (i *Info) ToRef() Ref {
	return Ref{ID: i.ID, Type: i.Type}
}

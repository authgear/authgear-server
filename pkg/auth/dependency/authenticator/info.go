package authenticator

import "github.com/authgear/authgear-server/pkg/core/authn"

type Info struct {
	ID            string                  `json:"id"`
	Type          authn.AuthenticatorType `json:"type"`
	Secret        string                  `json:"secret"`
	Props         map[string]interface{}  `json:"props"`
	Authenticator interface{}             `json:"-"`
}

func (i *Info) ToSpec() Spec {
	return Spec{Type: i.Type, Props: i.Props}
}

func (i *Info) ToRef() Ref {
	return Ref{ID: i.ID, Type: i.Type}
}

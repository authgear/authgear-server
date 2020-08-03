package authenticator

import "github.com/authgear/authgear-server/pkg/core/authn"

type Info struct {
	ID     string                  `json:"id"`
	UserID string                  `json:"user_id"`
	Type   authn.AuthenticatorType `json:"type"`
	Secret string                  `json:"secret"`
	Tag    []string                `json:"tag,omitempty"`
	Props  map[string]interface{}  `json:"props"`
}

func (i *Info) ToSpec() Spec {
	return Spec{
		UserID: i.UserID,
		Type:   i.Type,
		Tag:    i.Tag,
		Props:  i.Props,
	}
}

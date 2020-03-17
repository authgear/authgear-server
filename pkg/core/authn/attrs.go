package authn

import "time"

type Attrs struct {
	UserID string `json:"user_id"`

	PrincipalID        string        `json:"principal_id"`
	PrincipalType      PrincipalType `json:"principal_type"`
	PrincipalUpdatedAt time.Time     `json:"principal_updated_at"`

	AuthenticatorID         string                  `json:"authenticator_id,omitempty"`
	AuthenticatorType       AuthenticatorType       `json:"authenticator_type,omitempty"`
	AuthenticatorOOBChannel AuthenticatorOOBChannel `json:"authenticator_oob_channel,omitempty"`
	AuthenticatorUpdatedAt  *time.Time              `json:"authenticator_updated_at,omitempty"`
}

func (a *Attrs) AuthnAttrs() *Attrs {
	return a
}

type Attributer interface {
	AuthnAttrs() *Attrs
}

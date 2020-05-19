package authn

type Attrs struct {
	UserID string `json:"user_id"`

	ACR string   `json:"acr,omitempty"`
	AMR []string `json:"amr,omitempty"`
}

func (a *Attrs) AuthnAttrs() *Attrs {
	return a
}

type Attributer interface {
	AuthnAttrs() *Attrs
}

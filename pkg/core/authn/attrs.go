package authn

type Attrs struct {
	UserID string `json:"user_id"`

	IdentityType   IdentityType           `json:"identity_type"`
	IdentityClaims map[string]interface{} `json:"identity_claims"`
	ACR            string                 `json:"acr,omitempty"`
	AMR            []string               `json:"amr,omitempty"`
}

func (a *Attrs) AuthnAttrs() *Attrs {
	return a
}

type Attributer interface {
	AuthnAttrs() *Attrs
}

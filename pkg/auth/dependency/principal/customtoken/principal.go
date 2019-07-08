package customtoken

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/core/uuid"
)

type Principal struct {
	ID               string
	TokenPrincipalID string
	UserID           string
}

type attributes struct {
	ProviderUserID string      `json:"provider_user_id"`
	RawProfile     interface{} `json:"raw_profile"`
}

func NewPrincipal() Principal {
	return Principal{
		ID: uuid.New(),
	}
}

func (p *Principal) PrincipalID() string {
	return p.ID
}

func (p *Principal) PrincipalUserID() string {
	return p.UserID
}

func (p *Principal) ProviderType() string {
	return providerName
}

func (p *Principal) Attributes() principal.Attributes {
	return attributes{
		ProviderUserID: p.TokenPrincipalID,
		RawProfile:     struct{}{},
	}
}

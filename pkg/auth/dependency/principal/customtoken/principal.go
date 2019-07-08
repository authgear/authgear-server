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

func (p *Principal) ProviderID() string {
	return providerName
}

func (p *Principal) Attributes() principal.Attributes {
	return principal.Attributes{
		"provider_user_id": p.TokenPrincipalID,
		"raw_profile":      struct{}{},
	}
}

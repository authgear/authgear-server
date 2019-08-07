package anonymous

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/core/uuid"
)

type Principal struct {
	ID     string
	UserID string
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
	return providerAnonymous
}

func (p *Principal) Attributes() principal.Attributes {
	return principal.Attributes{}
}

func (p *Principal) Claims() principal.Claims {
	return principal.Claims{}
}

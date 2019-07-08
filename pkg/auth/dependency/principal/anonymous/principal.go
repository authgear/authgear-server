package anonymous

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/core/uuid"
)

type Principal struct {
	ID     string
	UserID string
}

type attributes struct {
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
	return providerAnonymous
}

func (p *Principal) Attributes() principal.Attributes {
	return attributes{}
}

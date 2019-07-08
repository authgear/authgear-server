package oauth

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"time"

	"github.com/skygeario/skygear-server/pkg/core/uuid"
)

type Principal struct {
	ID              string
	UserID          string
	ProviderName    string
	ProviderUserID  string
	AccessTokenResp interface{}
	UserProfile     interface{}
	CreatedAt       *time.Time
	UpdatedAt       *time.Time
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
		"provider_id":      p.ProviderName,
		"provider_user_id": p.ProviderUserID,
		"raw_profile":      p.UserProfile,
	}
}

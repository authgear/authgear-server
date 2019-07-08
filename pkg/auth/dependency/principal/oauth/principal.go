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

type attributes struct {
	ProviderID     string      `json:"provider_id"`
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
		ProviderID:     p.ProviderName,
		ProviderUserID: p.ProviderUserID,
		RawProfile:     p.UserProfile,
	}
}

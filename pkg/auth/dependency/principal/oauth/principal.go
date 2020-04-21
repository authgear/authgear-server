package oauth

import (
	"time"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	coreauthn "github.com/skygeario/skygear-server/pkg/core/authn"
	"github.com/skygeario/skygear-server/pkg/core/uuid"
)

type Principal struct {
	ID     string
	UserID string
	// (ProviderType, ProviderKeys, ProviderUserID) together form a unique index.
	ProviderType    string
	ProviderKeys    map[string]interface{}
	ProviderUserID  string
	AccessTokenResp interface{}
	UserProfile     interface{}
	ClaimsValue     map[string]interface{}
	CreatedAt       *time.Time
	UpdatedAt       *time.Time
}

func NewPrincipal(providerKeys map[string]interface{}) *Principal {
	if providerKeys == nil {
		providerKeys = map[string]interface{}{}
	}
	return &Principal{
		ID:           uuid.New(),
		ProviderKeys: providerKeys,
	}
}

func (p *Principal) PrincipalID() string {
	return p.ID
}

func (p *Principal) PrincipalUserID() string {
	return p.UserID
}

func (p *Principal) ProviderID() string {
	return string(coreauthn.PrincipalTypeOAuth)
}

func (p *Principal) Attributes() principal.Attributes {
	return principal.Attributes{
		"provider_type":    p.ProviderType,
		"provider_keys":    p.ProviderKeys,
		"provider_user_id": p.ProviderUserID,
		"raw_profile":      p.UserProfile,
	}
}

func (p *Principal) Claims() principal.Claims {
	providerID := map[string]interface{}{
		"type": p.ProviderType,
	}
	for k, v := range p.ProviderKeys {
		providerID[k] = v
	}

	claims := principal.Claims{
		"https://auth.skygear.io/claims/oauth/provider":   providerID,
		"https://auth.skygear.io/claims/oauth/subject_id": p.ProviderUserID,
		"https://auth.skygear.io/claims/oauth/profile":    p.UserProfile,
	}
	for k, v := range p.ClaimsValue {
		claims[k] = v
	}
	return claims
}

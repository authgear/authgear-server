package oauth

import (
	"time"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/config"
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
	return string(coreAuth.PrincipalTypeOAuth)
}

func (p *Principal) Attributes() principal.Attributes {
	// TODO: promote tenant
	return principal.Attributes{
		"provider_type":    p.ProviderType,
		"provider_user_id": p.ProviderUserID,
		"raw_profile":      p.UserProfile,
	}
}

func (p *Principal) Claims() principal.Claims {
	return p.ClaimsValue
}

func (p *Principal) SetRawProfile(rawProfile interface{}) {
	p.UserProfile = rawProfile
	rawProfileMap, ok := rawProfile.(map[string]interface{})
	if !ok {
		return
	}
	decoder := sso.GetUserInfoDecoder(config.OAuthProviderType(p.ProviderType))
	providerUserInfo := decoder.DecodeUserInfo(rawProfileMap)
	claimsValue := map[string]interface{}{}
	if providerUserInfo.Email != "" {
		claimsValue["email"] = providerUserInfo.Email
	}
	p.ClaimsValue = claimsValue
}

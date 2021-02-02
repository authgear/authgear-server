package sso

import (
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

// AuthInfo contains auth info from HandleAuthzResp
type AuthInfo struct {
	ProviderConfig          config.OAuthSSOProviderConfig
	ProviderRawProfile      map[string]interface{}
	ProviderAccessTokenResp interface{}
	ProviderUserInfo        ProviderUserInfo
}

type ProviderUserInfo struct {
	ID string
	// Email is normalized.
	Email string
}

func (i ProviderUserInfo) ClaimsValue() map[string]interface{} {
	claimsValue := map[string]interface{}{}
	if i.Email != "" {
		claimsValue[identity.StandardClaimEmail] = i.Email
	}
	return claimsValue
}

type OAuthAuthorizationResponse struct {
	Code  string
	State string
	Scope string
}

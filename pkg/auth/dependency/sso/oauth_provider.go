package sso

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/urlprefix"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

// OAuthProvider is OAuth 2.0 based provider.
type OAuthProvider interface {
	GetAuthURL(state State) (url string, err error)
	GetAuthInfo(r OAuthAuthorizationResponse) (AuthInfo, error)
}

// NonOpenIDConnectProvider are OAuth 2.0 provider that does not
// implement OpenID Connect or we do not implement yet.
// They are Google, Facebook, Instagram and LinkedIn.
type NonOpenIDConnectProvider interface {
	NonOpenIDConnectGetAuthInfo(r OAuthAuthorizationResponse) (authInfo AuthInfo, err error)
}

// ExternalAccessTokenFlowProvider is provider that the developer
// can somehow acquire an access token and that access token
// can be used to fetch user info.
// They are Google, Facebook, Instagram and LinkedIn.
type ExternalAccessTokenFlowProvider interface {
	ExternalAccessTokenGetAuthInfo(AccessTokenResp) (AuthInfo, error)
}

// OpenIDConnectProvider are OpenID Connect provider.
// They are Azure AD v2.
type OpenIDConnectProvider interface {
	OpenIDConnectGetAuthInfo(r OAuthAuthorizationResponse) (authInfo AuthInfo, err error)
}

type OAuthProviderFactory struct {
	urlPrefixProvider urlprefix.Provider
	tenantConfig      config.TenantConfiguration
}

func NewOAuthProviderFactory(tenantConfig config.TenantConfiguration, urlPrefixProvider urlprefix.Provider) *OAuthProviderFactory {
	return &OAuthProviderFactory{
		tenantConfig:      tenantConfig,
		urlPrefixProvider: urlPrefixProvider,
	}
}

func (p *OAuthProviderFactory) NewOAuthProvider(id string) OAuthProvider {
	providerConfig, ok := p.tenantConfig.GetOAuthProviderByID(id)
	if !ok {
		return nil
	}
	switch providerConfig.Type {
	case config.OAuthProviderTypeGoogle:
		return &GoogleImpl{
			URLPrefix:      p.urlPrefixProvider.Value(),
			OAuthConfig:    p.tenantConfig.UserConfig.SSO.OAuth,
			ProviderConfig: providerConfig,
		}
	case config.OAuthProviderTypeFacebook:
		return &FacebookImpl{
			URLPrefix:      p.urlPrefixProvider.Value(),
			OAuthConfig:    p.tenantConfig.UserConfig.SSO.OAuth,
			ProviderConfig: providerConfig,
		}
	case config.OAuthProviderTypeInstagram:
		return &InstagramImpl{
			URLPrefix:      p.urlPrefixProvider.Value(),
			OAuthConfig:    p.tenantConfig.UserConfig.SSO.OAuth,
			ProviderConfig: providerConfig,
		}
	case config.OAuthProviderTypeLinkedIn:
		return &LinkedInImpl{
			URLPrefix:      p.urlPrefixProvider.Value(),
			OAuthConfig:    p.tenantConfig.UserConfig.SSO.OAuth,
			ProviderConfig: providerConfig,
		}
	case config.OAuthProviderTypeAzureADv2:
		return &Azureadv2Impl{
			URLPrefix:      p.urlPrefixProvider.Value(),
			OAuthConfig:    p.tenantConfig.UserConfig.SSO.OAuth,
			ProviderConfig: providerConfig,
		}
	}
	return nil
}

func (p *OAuthProviderFactory) GetOAuthProviderConfig(id string) (config.OAuthProviderConfiguration, bool) {
	return p.tenantConfig.GetOAuthProviderByID(id)
}

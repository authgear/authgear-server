package sso

import (
	"net/url"

	"github.com/skygeario/skygear-server/pkg/clock"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

// OAuthProvider is OAuth 2.0 based provider.
type OAuthProvider interface {
	Type() config.OAuthProviderType
	GetAuthURL(state State, encodedState string) (url string, err error)
	GetAuthInfo(r OAuthAuthorizationResponse, state State) (AuthInfo, error)
}

// NonOpenIDConnectProvider are OAuth 2.0 provider that does not
// implement OpenID Connect or we do not implement yet.
// They are Google, Facebook and LinkedIn.
type NonOpenIDConnectProvider interface {
	NonOpenIDConnectGetAuthInfo(r OAuthAuthorizationResponse, state State) (authInfo AuthInfo, err error)
}

// ExternalAccessTokenFlowProvider is provider that the developer
// can somehow acquire an access token and that access token
// can be used to fetch user info.
// They are Facebook.
type ExternalAccessTokenFlowProvider interface {
	ExternalAccessTokenGetAuthInfo(AccessTokenResp) (AuthInfo, error)
}

// OpenIDConnectProvider are OpenID Connect provider.
// They are Azure AD v2.
type OpenIDConnectProvider interface {
	OpenIDConnectGetAuthInfo(r OAuthAuthorizationResponse, state State) (authInfo AuthInfo, err error)
}

type EndpointsProvider interface {
	BaseURL() *url.URL
}

type OAuthProviderFactory struct {
	Endpoints                EndpointsProvider
	redirectURIFunc          RedirectURLFunc
	tenantConfig             config.TenantConfiguration
	clock                    clock.Clock
	userInfoDecoder          UserInfoDecoder
	loginIDNormalizerFactory LoginIDNormalizerFactory
}

func NewOAuthProviderFactory(tenantConfig config.TenantConfiguration, endpoints EndpointsProvider, timeProvider clock.Clock, userInfoDecoder UserInfoDecoder, loginIDNormalizerFactory LoginIDNormalizerFactory, redirectURIFunc RedirectURLFunc) *OAuthProviderFactory {
	return &OAuthProviderFactory{
		tenantConfig:             tenantConfig,
		Endpoints:                endpoints,
		clock:                    timeProvider,
		userInfoDecoder:          userInfoDecoder,
		loginIDNormalizerFactory: loginIDNormalizerFactory,
		redirectURIFunc:          redirectURIFunc,
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
			URLPrefix:                p.Endpoints.BaseURL(),
			RedirectURLFunc:          p.redirectURIFunc,
			OAuthConfig:              p.tenantConfig.AppConfig.Identity.OAuth,
			ProviderConfig:           providerConfig,
			Clock:                    p.clock,
			UserInfoDecoder:          p.userInfoDecoder,
			LoginIDNormalizerFactory: p.loginIDNormalizerFactory,
		}
	case config.OAuthProviderTypeFacebook:
		return &FacebookImpl{
			URLPrefix:       p.Endpoints.BaseURL(),
			RedirectURLFunc: p.redirectURIFunc,
			OAuthConfig:     p.tenantConfig.AppConfig.Identity.OAuth,
			ProviderConfig:  providerConfig,
			UserInfoDecoder: p.userInfoDecoder,
		}
	case config.OAuthProviderTypeLinkedIn:
		return &LinkedInImpl{
			URLPrefix:       p.Endpoints.BaseURL(),
			RedirectURLFunc: p.redirectURIFunc,
			OAuthConfig:     p.tenantConfig.AppConfig.Identity.OAuth,
			ProviderConfig:  providerConfig,
			UserInfoDecoder: p.userInfoDecoder,
		}
	case config.OAuthProviderTypeAzureADv2:
		return &Azureadv2Impl{
			URLPrefix:                p.Endpoints.BaseURL(),
			RedirectURLFunc:          p.redirectURIFunc,
			OAuthConfig:              p.tenantConfig.AppConfig.Identity.OAuth,
			ProviderConfig:           providerConfig,
			Clock:                    p.clock,
			LoginIDNormalizerFactory: p.loginIDNormalizerFactory,
		}
	case config.OAuthProviderTypeApple:
		return &AppleImpl{
			URLPrefix:                p.Endpoints.BaseURL(),
			RedirectURLFunc:          p.redirectURIFunc,
			OAuthConfig:              p.tenantConfig.AppConfig.Identity.OAuth,
			ProviderConfig:           providerConfig,
			Clock:                    p.clock,
			LoginIDNormalizerFactory: p.loginIDNormalizerFactory,
		}
	}
	return nil
}

func (p *OAuthProviderFactory) GetOAuthProviderConfig(id string) (config.OAuthProviderConfiguration, bool) {
	return p.tenantConfig.GetOAuthProviderByID(id)
}

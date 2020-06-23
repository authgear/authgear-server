package sso

import (
	"net/url"

	"github.com/skygeario/skygear-server/pkg/auth/config"
	"github.com/skygeario/skygear-server/pkg/clock"
)

// OAuthProvider is OAuth 2.0 based provider.
type OAuthProvider interface {
	Type() config.OAuthSSOProviderType
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
	IdentityConfig           *config.IdentityConfig
	Credentials              *config.OAuthClientCredentials
	RedirectURIFunc          RedirectURLFunc
	Clock                    clock.Clock
	UserInfoDecoder          UserInfoDecoder
	LoginIDNormalizerFactory LoginIDNormalizerFactory
}

func (p *OAuthProviderFactory) NewOAuthProvider(id string) OAuthProvider {
	providerConfig, ok := p.IdentityConfig.OAuth.GetProviderConfig(id)
	if !ok {
		return nil
	}
	credentials, ok := p.Credentials.Lookup(id)
	if !ok {
		return nil
	}

	switch providerConfig.Type {
	case config.OAuthSSOProviderTypeGoogle:
		return &GoogleImpl{
			URLPrefix:                p.Endpoints.BaseURL(),
			RedirectURLFunc:          p.RedirectURIFunc,
			ProviderConfig:           *providerConfig,
			Credentials:              *credentials,
			Clock:                    p.Clock,
			LoginIDNormalizerFactory: p.LoginIDNormalizerFactory,
		}
	case config.OAuthSSOProviderTypeFacebook:
		return &FacebookImpl{
			URLPrefix:       p.Endpoints.BaseURL(),
			RedirectURLFunc: p.RedirectURIFunc,
			ProviderConfig:  *providerConfig,
			Credentials:     *credentials,
			UserInfoDecoder: p.UserInfoDecoder,
		}
	case config.OAuthSSOProviderTypeLinkedIn:
		return &LinkedInImpl{
			URLPrefix:       p.Endpoints.BaseURL(),
			RedirectURLFunc: p.RedirectURIFunc,
			ProviderConfig:  *providerConfig,
			Credentials:     *credentials,
			UserInfoDecoder: p.UserInfoDecoder,
		}
	case config.OAuthSSOProviderTypeAzureADv2:
		return &Azureadv2Impl{
			URLPrefix:                p.Endpoints.BaseURL(),
			RedirectURLFunc:          p.RedirectURIFunc,
			ProviderConfig:           *providerConfig,
			Credentials:              *credentials,
			Clock:                    p.Clock,
			LoginIDNormalizerFactory: p.LoginIDNormalizerFactory,
		}
	case config.OAuthSSOProviderTypeApple:
		return &AppleImpl{
			URLPrefix:                p.Endpoints.BaseURL(),
			RedirectURLFunc:          p.RedirectURIFunc,
			ProviderConfig:           *providerConfig,
			Credentials:              *credentials,
			Clock:                    p.Clock,
			LoginIDNormalizerFactory: p.LoginIDNormalizerFactory,
		}
	}
	return nil
}

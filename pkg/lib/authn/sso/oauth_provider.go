package sso

import (
	"net/url"

	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

type GetAuthURLParam struct {
	Nonce  string
	State  string
	Prompt []string
}

type GetAuthInfoParam struct {
	Nonce string
}

type OAuthAuthorizationResponse struct {
	Code  string
	State string
	Scope string
}

// OAuthProvider is OAuth 2.0 based provider.
type OAuthProvider interface {
	Type() config.OAuthSSOProviderType
	Config() config.OAuthSSOProviderConfig
	GetAuthURL(param GetAuthURLParam) (url string, err error)
	GetAuthInfo(r OAuthAuthorizationResponse, param GetAuthInfoParam) (AuthInfo, error)
	GetPrompt(prompt []string) []string
}

// NonOpenIDConnectProvider are OAuth 2.0 provider that does not
// implement OpenID Connect or we do not implement yet.
// They are
// "facebook"
// "linkedin"
type NonOpenIDConnectProvider interface {
	NonOpenIDConnectGetAuthInfo(r OAuthAuthorizationResponse, param GetAuthInfoParam) (authInfo AuthInfo, err error)
}

// OpenIDConnectProvider are OpenID Connect provider.
// They are
// "google"
// "apple"
// "azureadv2"
// "adfs"
type OpenIDConnectProvider interface {
	OpenIDConnectGetAuthInfo(r OAuthAuthorizationResponse, param GetAuthInfoParam) (authInfo AuthInfo, err error)
}

type EndpointsProvider interface {
	BaseURL() *url.URL
}

type RedirectURLProvider interface {
	SSOCallbackURL(providerConfig config.OAuthSSOProviderConfig) *url.URL
}

type StandardAttributesNormalizer interface {
	Normalize(stdattrs.T) error
}

type OAuthProviderFactory struct {
	Endpoints                    EndpointsProvider
	IdentityConfig               *config.IdentityConfig
	Credentials                  *config.OAuthClientCredentials
	RedirectURL                  RedirectURLProvider
	Clock                        clock.Clock
	WechatURLProvider            WechatURLProvider
	StandardAttributesNormalizer StandardAttributesNormalizer
}

func (p *OAuthProviderFactory) NewOAuthProvider(alias string) OAuthProvider {
	providerConfig, ok := p.IdentityConfig.OAuth.GetProviderConfig(alias)
	if !ok {
		return nil
	}
	credentials, ok := p.Credentials.Lookup(alias)
	if !ok {
		return nil
	}

	switch providerConfig.Type {
	case config.OAuthSSOProviderTypeGoogle:
		return &GoogleImpl{
			Clock:                        p.Clock,
			RedirectURL:                  p.RedirectURL,
			ProviderConfig:               *providerConfig,
			Credentials:                  *credentials,
			StandardAttributesNormalizer: p.StandardAttributesNormalizer,
		}
	case config.OAuthSSOProviderTypeFacebook:
		return &FacebookImpl{
			RedirectURL:                  p.RedirectURL,
			ProviderConfig:               *providerConfig,
			Credentials:                  *credentials,
			StandardAttributesNormalizer: p.StandardAttributesNormalizer,
		}
	case config.OAuthSSOProviderTypeLinkedIn:
		return &LinkedInImpl{
			RedirectURL:                  p.RedirectURL,
			ProviderConfig:               *providerConfig,
			Credentials:                  *credentials,
			StandardAttributesNormalizer: p.StandardAttributesNormalizer,
		}
	case config.OAuthSSOProviderTypeAzureADv2:
		return &Azureadv2Impl{
			Clock:                        p.Clock,
			RedirectURL:                  p.RedirectURL,
			ProviderConfig:               *providerConfig,
			Credentials:                  *credentials,
			StandardAttributesNormalizer: p.StandardAttributesNormalizer,
		}
	case config.OAuthSSOProviderTypeADFS:
		return &ADFSImpl{
			Clock:                        p.Clock,
			RedirectURL:                  p.RedirectURL,
			ProviderConfig:               *providerConfig,
			Credentials:                  *credentials,
			StandardAttributesNormalizer: p.StandardAttributesNormalizer,
		}
	case config.OAuthSSOProviderTypeApple:
		return &AppleImpl{
			Clock:                        p.Clock,
			RedirectURL:                  p.RedirectURL,
			ProviderConfig:               *providerConfig,
			Credentials:                  *credentials,
			StandardAttributesNormalizer: p.StandardAttributesNormalizer,
		}
	case config.OAuthSSOProviderTypeWechat:
		return &WechatImpl{
			ProviderConfig:               *providerConfig,
			Credentials:                  *credentials,
			URLProvider:                  p.WechatURLProvider,
			StandardAttributesNormalizer: p.StandardAttributesNormalizer,
		}
	}
	return nil
}

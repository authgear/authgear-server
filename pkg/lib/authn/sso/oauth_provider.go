package sso

import (
	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

type GetAuthURLParam struct {
	RedirectURI  string
	ResponseMode ResponseMode
	Nonce        string
	State        string
	Prompt       []string
}

type GetAuthInfoParam struct {
	RedirectURI string
	Nonce       string
}

type OAuthAuthorizationResponse struct {
	Code string
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
// "wechat"
type NonOpenIDConnectProvider interface {
	NonOpenIDConnectGetAuthInfo(r OAuthAuthorizationResponse, param GetAuthInfoParam) (authInfo AuthInfo, err error)
}

// OpenIDConnectProvider are OpenID Connect provider.
// They are
// "google"
// "apple"
// "azureadv2"
// "azureadb2c"
// "adfs"
type OpenIDConnectProvider interface {
	OpenIDConnectGetAuthInfo(r OAuthAuthorizationResponse, param GetAuthInfoParam) (authInfo AuthInfo, err error)
}

type StandardAttributesNormalizer interface {
	Normalize(stdattrs.T) error
}

type OAuthProviderFactory struct {
	IdentityConfig               *config.IdentityConfig
	Credentials                  *config.OAuthSSOProviderCredentials
	Clock                        clock.Clock
	StandardAttributesNormalizer StandardAttributesNormalizer
	HTTPClient                   OAuthHTTPClient
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
			ProviderConfig:               *providerConfig,
			Credentials:                  *credentials,
			StandardAttributesNormalizer: p.StandardAttributesNormalizer,
			HTTPClient:                   p.HTTPClient,
		}
	case config.OAuthSSOProviderTypeFacebook:
		return &FacebookImpl{
			ProviderConfig:               *providerConfig,
			Credentials:                  *credentials,
			StandardAttributesNormalizer: p.StandardAttributesNormalizer,
			HTTPClient:                   p.HTTPClient,
		}
	case config.OAuthSSOProviderTypeGithub:
		return &GithubImpl{
			ProviderConfig:               *providerConfig,
			Credentials:                  *credentials,
			StandardAttributesNormalizer: p.StandardAttributesNormalizer,
			HTTPClient:                   p.HTTPClient,
		}
	case config.OAuthSSOProviderTypeLinkedIn:
		return &LinkedInImpl{
			ProviderConfig:               *providerConfig,
			Credentials:                  *credentials,
			StandardAttributesNormalizer: p.StandardAttributesNormalizer,
			HTTPClient:                   p.HTTPClient,
		}
	case config.OAuthSSOProviderTypeAzureADv2:
		return &Azureadv2Impl{
			Clock:                        p.Clock,
			ProviderConfig:               *providerConfig,
			Credentials:                  *credentials,
			StandardAttributesNormalizer: p.StandardAttributesNormalizer,
			HTTPClient:                   p.HTTPClient,
		}
	case config.OAuthSSOProviderTypeAzureADB2C:
		return &Azureadb2cImpl{
			Clock:                        p.Clock,
			ProviderConfig:               *providerConfig,
			Credentials:                  *credentials,
			StandardAttributesNormalizer: p.StandardAttributesNormalizer,
			HTTPClient:                   p.HTTPClient,
		}
	case config.OAuthSSOProviderTypeADFS:
		return &ADFSImpl{
			Clock:                        p.Clock,
			ProviderConfig:               *providerConfig,
			Credentials:                  *credentials,
			StandardAttributesNormalizer: p.StandardAttributesNormalizer,
			HTTPClient:                   p.HTTPClient,
		}
	case config.OAuthSSOProviderTypeApple:
		return &AppleImpl{
			Clock:                        p.Clock,
			ProviderConfig:               *providerConfig,
			Credentials:                  *credentials,
			StandardAttributesNormalizer: p.StandardAttributesNormalizer,
			HTTPClient:                   p.HTTPClient,
		}
	case config.OAuthSSOProviderTypeWechat:
		return &WechatImpl{
			ProviderConfig:               *providerConfig,
			Credentials:                  *credentials,
			StandardAttributesNormalizer: p.StandardAttributesNormalizer,
			HTTPClient:                   p.HTTPClient,
		}
	}
	return nil
}

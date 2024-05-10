package sso

import (
	"github.com/authgear/authgear-server/pkg/api/oauthrelyingparty"
	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/adfs"
	"github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/apple"
	"github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/azureadb2c"
	"github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/azureadv2"
	"github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/facebook"
	"github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/github"
	"github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/google"
	"github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/linkedin"
	"github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/wechat"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

type GetAuthURLParam struct {
	RedirectURI  string
	ResponseMode string
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
	Config() oauthrelyingparty.ProviderConfig
	GetAuthURL(param GetAuthURLParam) (url string, err error)
	GetAuthInfo(r OAuthAuthorizationResponse, param GetAuthInfoParam) (AuthInfo, error)
	GetPrompt(prompt []string) []string
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

	switch providerConfig.Type() {
	case google.Type:
		return &GoogleImpl{
			Clock:                        p.Clock,
			ProviderConfig:               *providerConfig,
			ClientSecret:                 credentials.ClientSecret,
			StandardAttributesNormalizer: p.StandardAttributesNormalizer,
			HTTPClient:                   p.HTTPClient,
		}
	case facebook.Type:
		return &FacebookImpl{
			ProviderConfig:               *providerConfig,
			ClientSecret:                 credentials.ClientSecret,
			StandardAttributesNormalizer: p.StandardAttributesNormalizer,
			HTTPClient:                   p.HTTPClient,
		}
	case github.Type:
		return &GithubImpl{
			ProviderConfig:               *providerConfig,
			ClientSecret:                 credentials.ClientSecret,
			StandardAttributesNormalizer: p.StandardAttributesNormalizer,
			HTTPClient:                   p.HTTPClient,
		}
	case linkedin.Type:
		return &LinkedInImpl{
			ProviderConfig:               *providerConfig,
			ClientSecret:                 credentials.ClientSecret,
			StandardAttributesNormalizer: p.StandardAttributesNormalizer,
			HTTPClient:                   p.HTTPClient,
		}
	case azureadv2.Type:
		return &Azureadv2Impl{
			Clock:                        p.Clock,
			ProviderConfig:               *providerConfig,
			ClientSecret:                 credentials.ClientSecret,
			StandardAttributesNormalizer: p.StandardAttributesNormalizer,
			HTTPClient:                   p.HTTPClient,
		}
	case azureadb2c.Type:
		return &Azureadb2cImpl{
			Clock:                        p.Clock,
			ProviderConfig:               *providerConfig,
			ClientSecret:                 credentials.ClientSecret,
			StandardAttributesNormalizer: p.StandardAttributesNormalizer,
			HTTPClient:                   p.HTTPClient,
		}
	case adfs.Type:
		return &ADFSImpl{
			Clock:                        p.Clock,
			ProviderConfig:               *providerConfig,
			ClientSecret:                 credentials.ClientSecret,
			StandardAttributesNormalizer: p.StandardAttributesNormalizer,
			HTTPClient:                   p.HTTPClient,
		}
	case apple.Type:
		return &AppleImpl{
			Clock:                        p.Clock,
			ProviderConfig:               *providerConfig,
			ClientSecret:                 credentials.ClientSecret,
			StandardAttributesNormalizer: p.StandardAttributesNormalizer,
			HTTPClient:                   p.HTTPClient,
		}
	case wechat.Type:
		return &WechatImpl{
			ProviderConfig:               *providerConfig,
			ClientSecret:                 credentials.ClientSecret,
			StandardAttributesNormalizer: p.StandardAttributesNormalizer,
			HTTPClient:                   p.HTTPClient,
		}
	}
	return nil
}

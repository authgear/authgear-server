package sso

import (
	"github.com/authgear/authgear-server/pkg/api"
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

// OAuthProvider is OAuth 2.0 based provider.
type OAuthProvider interface {
	Config() oauthrelyingparty.ProviderConfig
	GetAuthorizationURL(options oauthrelyingparty.GetAuthorizationURLOptions) (url string, err error)
	GetUserProfile(options oauthrelyingparty.GetUserProfileOptions) (oauthrelyingparty.UserProfile, error)
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

func (p *OAuthProviderFactory) GetProviderConfig(alias string) (oauthrelyingparty.ProviderConfig, error) {
	providerConfig, ok := p.IdentityConfig.OAuth.GetProviderConfig(alias)
	if !ok {
		return nil, api.ErrOAuthProviderNotFound
	}
	return providerConfig, nil
}

func (p *OAuthProviderFactory) getProvider(alias string) (provider OAuthProvider, err error) {
	providerConfig, err := p.GetProviderConfig(alias)
	if err != nil {
		return
	}

	credentials, ok := p.Credentials.Lookup(alias)
	if !ok {
		err = api.ErrOAuthProviderNotFound
		return
	}

	switch providerConfig.Type() {
	case google.Type:
		return &GoogleImpl{
			Clock:                        p.Clock,
			ProviderConfig:               providerConfig,
			ClientSecret:                 credentials.ClientSecret,
			StandardAttributesNormalizer: p.StandardAttributesNormalizer,
			HTTPClient:                   p.HTTPClient,
		}, nil
	case facebook.Type:
		return &FacebookImpl{
			ProviderConfig:               providerConfig,
			ClientSecret:                 credentials.ClientSecret,
			StandardAttributesNormalizer: p.StandardAttributesNormalizer,
			HTTPClient:                   p.HTTPClient,
		}, nil
	case github.Type:
		return &GithubImpl{
			ProviderConfig:               providerConfig,
			ClientSecret:                 credentials.ClientSecret,
			StandardAttributesNormalizer: p.StandardAttributesNormalizer,
			HTTPClient:                   p.HTTPClient,
		}, nil
	case linkedin.Type:
		return &LinkedInImpl{
			ProviderConfig:               providerConfig,
			ClientSecret:                 credentials.ClientSecret,
			StandardAttributesNormalizer: p.StandardAttributesNormalizer,
			HTTPClient:                   p.HTTPClient,
		}, nil
	case azureadv2.Type:
		return &Azureadv2Impl{
			Clock:                        p.Clock,
			ProviderConfig:               providerConfig,
			ClientSecret:                 credentials.ClientSecret,
			StandardAttributesNormalizer: p.StandardAttributesNormalizer,
			HTTPClient:                   p.HTTPClient,
		}, nil
	case azureadb2c.Type:
		return &Azureadb2cImpl{
			Clock:                        p.Clock,
			ProviderConfig:               providerConfig,
			ClientSecret:                 credentials.ClientSecret,
			StandardAttributesNormalizer: p.StandardAttributesNormalizer,
			HTTPClient:                   p.HTTPClient,
		}, nil
	case adfs.Type:
		return &ADFSImpl{
			Clock:                        p.Clock,
			ProviderConfig:               providerConfig,
			ClientSecret:                 credentials.ClientSecret,
			StandardAttributesNormalizer: p.StandardAttributesNormalizer,
			HTTPClient:                   p.HTTPClient,
		}, nil
	case apple.Type:
		return &AppleImpl{
			Clock:                        p.Clock,
			ProviderConfig:               providerConfig,
			ClientSecret:                 credentials.ClientSecret,
			StandardAttributesNormalizer: p.StandardAttributesNormalizer,
			HTTPClient:                   p.HTTPClient,
		}, nil
	case wechat.Type:
		return &WechatImpl{
			ProviderConfig:               providerConfig,
			ClientSecret:                 credentials.ClientSecret,
			StandardAttributesNormalizer: p.StandardAttributesNormalizer,
			HTTPClient:                   p.HTTPClient,
		}, nil
	default:
		// TODO(oauth): switch to registry-based resolution.
		err = api.ErrOAuthProviderNotFound
		return
	}
}

func (p *OAuthProviderFactory) GetAuthorizationURL(alias string, options oauthrelyingparty.GetAuthorizationURLOptions) (url string, err error) {
	provider, err := p.getProvider(alias)
	if err != nil {
		return
	}

	return provider.GetAuthorizationURL(options)
}

func (p *OAuthProviderFactory) GetUserProfile(alias string, options oauthrelyingparty.GetUserProfileOptions) (userProfile oauthrelyingparty.UserProfile, err error) {
	provider, err := p.getProvider(alias)
	if err != nil {
		return
	}

	return provider.GetUserProfile(options)
}

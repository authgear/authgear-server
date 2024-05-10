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
	GetAuthorizationURL(deps oauthrelyingparty.Dependencies, options oauthrelyingparty.GetAuthorizationURLOptions) (url string, err error)
	GetUserProfile(deps oauthrelyingparty.Dependencies, options oauthrelyingparty.GetUserProfileOptions) (oauthrelyingparty.UserProfile, error)
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

func (p *OAuthProviderFactory) getProvider(alias string) (provider OAuthProvider, deps *oauthrelyingparty.Dependencies, err error) {
	providerConfig, err := p.GetProviderConfig(alias)
	if err != nil {
		return
	}

	credentials, ok := p.Credentials.Lookup(alias)
	if !ok {
		err = api.ErrOAuthProviderNotFound
		return
	}

	deps = &oauthrelyingparty.Dependencies{
		Clock:          p.Clock,
		ProviderConfig: providerConfig,
		ClientSecret:   credentials.ClientSecret,
		HTTPClient:     p.HTTPClient.Client,
	}

	switch providerConfig.Type() {
	case google.Type:
		provider = &GoogleImpl{}
		return
	case facebook.Type:
		provider = &FacebookImpl{}
		return
	case github.Type:
		provider = &GithubImpl{}
		return
	case linkedin.Type:
		provider = &LinkedInImpl{}
		return
	case azureadv2.Type:
		provider = &Azureadv2Impl{}
		return
	case azureadb2c.Type:
		provider = &Azureadb2cImpl{}
		return
	case adfs.Type:
		provider = &ADFSImpl{}
		return
	case apple.Type:
		provider = &AppleImpl{}
		return
	case wechat.Type:
		provider = &WechatImpl{}
		return
	default:
		// TODO(oauth): switch to registry-based resolution.
		err = api.ErrOAuthProviderNotFound
		return
	}
}

func (p *OAuthProviderFactory) GetAuthorizationURL(alias string, options oauthrelyingparty.GetAuthorizationURLOptions) (url string, err error) {
	provider, deps, err := p.getProvider(alias)
	if err != nil {
		return
	}

	return provider.GetAuthorizationURL(*deps, options)
}

func (p *OAuthProviderFactory) GetUserProfile(alias string, options oauthrelyingparty.GetUserProfileOptions) (userProfile oauthrelyingparty.UserProfile, err error) {
	provider, deps, err := p.getProvider(alias)
	if err != nil {
		return
	}

	userProfile, err = provider.GetUserProfile(*deps, options)
	if err != nil {
		return
	}

	err = p.StandardAttributesNormalizer.Normalize(userProfile.StandardAttributes)
	if err != nil {
		return
	}

	return
}

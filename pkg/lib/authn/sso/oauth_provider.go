package sso

import (
	"errors"

	"github.com/authgear/oauthrelyingparty/pkg/api/oauthrelyingparty"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/oauthrelyingpartyutil"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

type StandardAttributesNormalizer interface {
	Normalize(stdattrs.T) error
}

type OAuthProviderFactory struct {
	IdentityConfig               *config.IdentityConfig
	Credentials                  *config.OAuthSSOProviderCredentials
	Clock                        clock.Clock
	StandardAttributesNormalizer StandardAttributesNormalizer
	HTTPClient                   OAuthHTTPClient
	SimpleStoreRedisFactory      *SimpleStoreRedisFactory
}

func (p *OAuthProviderFactory) GetProviderConfig(alias string) (oauthrelyingparty.ProviderConfig, error) {
	providerConfig, ok := p.IdentityConfig.OAuth.GetProviderConfig(alias)
	if !ok {
		return nil, api.ErrOAuthProviderNotFound
	}
	return providerConfig, nil
}

func (p *OAuthProviderFactory) getProvider(alias string) (provider oauthrelyingparty.Provider, deps *oauthrelyingparty.Dependencies, err error) {
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
		SimpleStore:    p.SimpleStoreRedisFactory.GetStoreByProvider(providerConfig.Type(), alias),
	}

	provider = providerConfig.MustGetProvider()
	return
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
		var oauthErrorResponse *oauthrelyingparty.ErrorResponse
		if errors.As(err, &oauthErrorResponse) {
			err = oauthrelyingpartyutil.NewOAuthError(oauthErrorResponse)
			return
		}

		return
	}

	err = p.StandardAttributesNormalizer.Normalize(userProfile.StandardAttributes)
	if err != nil {
		return
	}

	return
}

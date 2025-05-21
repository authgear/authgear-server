package sso

import (
	"context"
	"errors"

	"github.com/authgear/oauthrelyingparty/pkg/api/oauthrelyingparty"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/oauthrelyingpartyutil"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

type StandardAttributesNormalizer interface {
	Normalize(context.Context, stdattrs.T) error
}

type OAuthProviderFactory struct {
	IdentityConfig               *config.IdentityConfig
	Credentials                  *config.OAuthSSOProviderCredentials
	SSOOAuthDemoCredentials      *config.SSOOAuthDemoCredentials
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

func (p *OAuthProviderFactory) getActiveOrDemoProvider(alias string) (provider oauthrelyingparty.Provider, deps *oauthrelyingparty.Dependencies, err error) {
	providerConfig, err := p.GetProviderConfig(alias)
	if err != nil {
		return
	}

	if config.OAuthSSOProviderConfig(providerConfig).IsMissingCredentialAllowed() {
		if p.SSOOAuthDemoCredentials == nil {
			err = newOAuthProviderMissingCredentialsError(alias, providerConfig.Type(), false)
			return
		}

		demoItem, ok := p.SSOOAuthDemoCredentials.LookupByProviderType(providerConfig.Type())
		if !ok {
			err = newOAuthProviderMissingCredentialsError(alias, providerConfig.Type(), true)
			return
		}

		deps = &oauthrelyingparty.Dependencies{
			Clock:          p.Clock,
			ProviderConfig: demoItem.ProviderConfig,
			ClientSecret:   demoItem.ClientSecret,
			HTTPClient:     p.HTTPClient.Client,
			SimpleStore:    p.SimpleStoreRedisFactory.GetStoreByProvider(providerConfig.Type(), alias),
		}

		provider = demoItem.ProviderConfig.MustGetProvider()
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

func (p *OAuthProviderFactory) GetAuthorizationURL(ctx context.Context, alias string, options oauthrelyingparty.GetAuthorizationURLOptions) (url string, err error) {
	provider, deps, err := p.getActiveOrDemoProvider(alias)
	if err != nil {
		return
	}

	return provider.GetAuthorizationURL(ctx, *deps, options)
}

func (p *OAuthProviderFactory) GetUserProfile(ctx context.Context, alias string, options oauthrelyingparty.GetUserProfileOptions) (userProfile oauthrelyingparty.UserProfile, err error) {
	provider, deps, err := p.getActiveOrDemoProvider(alias)
	if err != nil {
		return
	}

	userProfile, err = provider.GetUserProfile(ctx, *deps, options)
	if err != nil {
		var oauthErrorResponse *oauthrelyingparty.ErrorResponse
		if errors.As(err, &oauthErrorResponse) {
			err = oauthrelyingpartyutil.NewOAuthError(oauthErrorResponse)
			return
		}

		return
	}

	err = p.StandardAttributesNormalizer.Normalize(ctx, userProfile.StandardAttributes)
	if err != nil {
		return
	}

	return
}

func newOAuthProviderMissingCredentialsError(alias string, providerType string, isDemo bool) error {
	details := apierrors.Details{
		"OAuthProviderAlias": alias,
		"OAuthProviderType":  providerType,
	}
	return api.OAuthProviderMissingCredentials.NewWithInfo("oauth provider is missing credentials", details)
}

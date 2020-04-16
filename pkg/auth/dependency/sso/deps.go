package sso

import (
	"context"

	"github.com/google/wire"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/loginid"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/urlprefix"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/time"
)

func ProvideSSOProvider(
	ctx context.Context,
	config *config.TenantConfiguration,
	store SkygearAuthorizationCodeStore,
) Provider {
	return NewProvider(
		ctx,
		config.AppID,
		config.AppConfig.Identity.OAuth,
		store,
	)
}

func ProvideSkygearAuthorizationCodeStore(ctx context.Context) SkygearAuthorizationCodeStore {
	return NewSkygearAuthorizationCodeStore(ctx)
}

func ProvideOAuthProviderFactory(
	cfg *config.TenantConfiguration,
	up urlprefix.Provider,
	tp time.Provider,
	nf loginid.LoginIDNormalizerFactory,
	rf RedirectURLFunc,
) *OAuthProviderFactory {
	return NewOAuthProviderFactory(*cfg, up, tp, NewUserInfoDecoder(nf), nf, rf)
}

func ProvideAuthHandlerHTMLProvider(up urlprefix.Provider) AuthHandlerHTMLProvider {
	return NewAuthHandlerHTMLProvider(up.Value())
}

var DependencySet = wire.NewSet(
	ProvideSSOProvider,
	ProvideSkygearAuthorizationCodeStore,
	ProvideOAuthProviderFactory,
	ProvideAuthHandlerHTMLProvider,
)

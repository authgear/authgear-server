package sso

import (
	"github.com/google/wire"

	"github.com/skygeario/skygear-server/pkg/clock"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

func ProvideStateCodec(config *config.TenantConfiguration) *StateCodec {
	return NewStateCodec(
		config.AppID,
		config.AppConfig.Identity.OAuth,
	)
}

func ProvideOAuthProviderFactory(
	cfg *config.TenantConfiguration,
	ep EndpointsProvider,
	tp clock.Clock,
	nf LoginIDNormalizerFactory,
	rf RedirectURLFunc,
) *OAuthProviderFactory {
	return NewOAuthProviderFactory(*cfg, ep, tp, NewUserInfoDecoder(nf), nf, rf)
}

var DependencySet = wire.NewSet(
	ProvideStateCodec,
	ProvideOAuthProviderFactory,
)

package handler

import (
	"github.com/google/wire"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oidc"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

func ProvideEndSessionHandler(
	cfg *config.TenantConfiguration,
	endSession oidc.EndSessionEndpointProvider,
	logout oauth.LogoutEndpointProvider,
	settings oauth.SettingsEndpointProvider,
) *EndSessionHandler {
	return &EndSessionHandler{
		Clients:            cfg.AppConfig.Clients,
		EndSessionEndpoint: endSession,
		LogoutEndpoint:     logout,
		SettingsEndpoint:   settings,
	}
}

var DependencySet = wire.NewSet(
	ProvideEndSessionHandler,
)

package handler

import (
	"github.com/google/wire"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oidc"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

func ProvideEndSessionHandler(
	cfg *config.TenantConfiguration,
	endSession oidc.EndSessionEndpointProvider,
	logout LogoutURLProvider,
	settings SettingsURLProvider,
) *EndSessionHandler {
	return &EndSessionHandler{
		Clients:            cfg.AppConfig.Clients,
		EndSessionEndpoint: endSession,
		LogoutURL:          logout,
		SettingsURL:        settings,
	}
}

var DependencySet = wire.NewSet(
	ProvideEndSessionHandler,
)

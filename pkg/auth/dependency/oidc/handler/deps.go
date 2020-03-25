package handler

import (
	"github.com/google/wire"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

func ProvideEndSessionHandler(
	cfg *config.TenantConfiguration,
	sm *auth.SessionManager,
	settings oauth.SettingsEndpointProvider,
) *EndSessionHandler {
	return &EndSessionHandler{
		Clients:          cfg.AppConfig.Clients,
		Sessions:         sm,
		SettingsEndpoint: settings,
	}
}

var DependencySet = wire.NewSet(
	ProvideEndSessionHandler,
)

package oauth

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(MetadataProvider), "*"),
	wire.Struct(new(Resolver), "*"),
	wire.Struct(new(SessionManager), "*"),
	wire.Struct(new(OfflineGrantService), "*"),
	wire.Struct(new(PromptResolver), "*"),

	wire.Struct(new(AccessTokenEncoding), "*"),
	wire.Bind(new(AccessTokenDecoder), new(*AccessTokenEncoding)),
	wire.Struct(new(AuthorizationService), "*"),
	wire.Bind(new(OfflineGrantSessionManager), new(*SessionManager)),

	wire.Struct(new(AppSessionTokenService), "*"),
	wire.Bind(new(AppSessionTokenServiceOfflineGrantService), new(*OfflineGrantService)),

	wire.Struct(new(AppInitiatedSSOToWebTokenService), "*"),
)

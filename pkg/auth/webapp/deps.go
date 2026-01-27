package webapp

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(AnonymousUserPromotionService), "*"),

	NewSessionCookieDef,
	NewErrorTokenCookieDef,
	NewSignedUpCookieDef,
	wire.Struct(new(ErrorService), "*"),
	wire.Struct(new(AuthflowNavigator), "*"),

	wire.Struct(new(SessionMiddleware), "*"),
	wire.Bind(new(SessionMiddlewareStore), new(*SessionStoreRedis)),
	wire.Bind(new(SessionMiddlewareSessionService), new(*Service2)),
	wire.Struct(new(ColorSchemeMiddleware), "*"),
	wire.Struct(new(WeChatRedirectURIMiddleware), "*"),
	wire.Struct(new(UIParamMiddleware), "*"),
	wire.Struct(new(VisitorIDMiddleware), "*"),
	wire.Struct(new(RequireAuthenticationEnabledMiddleware), "*"),
	wire.Struct(new(RequireSettingsEnabledMiddleware), "*"),
	wire.Struct(new(SettingsSubRoutesMiddleware), "*"),
	wire.Struct(new(SuccessPageMiddleware), "*"),
	wire.Struct(new(TutorialMiddleware), "*"),
	wire.Struct(new(DynamicCSPMiddleware), "*"),
	wire.Struct(new(ContextHolderMiddleware), "*"),

	wire.Struct(new(PublicOriginMiddleware), "*"),

	wire.Struct(new(SessionStoreRedis), "*"),
	wire.Bind(new(SessionStore), new(*SessionStoreRedis)),
	wire.Struct(new(Service2), "*"),
)

var RootMiddlewareDependencySet = wire.NewSet(
	wire.Struct(new(NoProjectCSPMiddleware), "*"),
	wire.Struct(new(ContextHolderMiddleware), "*"),
)

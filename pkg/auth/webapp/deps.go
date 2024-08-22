package webapp

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(AnonymousUserPromotionService), "*"),

	NewCSRFCookieDef,
	NewSessionCookieDef,
	NewErrorCookieDef,
	NewSignedUpCookieDef,
	wire.Struct(new(ErrorCookie), "*"),

	wire.Struct(new(CSRFDebugMiddleware), "*"),
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
	NewContextHolderMiddlewareLogger,
	wire.Struct(new(ContextHolderMiddleware), "*"),

	NewPublicOriginMiddlewareLogger,
	wire.Struct(new(PublicOriginMiddleware), "*"),

	NewServiceLogger,
	wire.Struct(new(SessionStoreRedis), "*"),
	wire.Bind(new(SessionStore), new(*SessionStoreRedis)),
	wire.Struct(new(Service2), "*"),
)

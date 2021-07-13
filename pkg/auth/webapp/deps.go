package webapp

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(URLProvider), "*"),
	wire.Struct(new(AuthenticateURLProvider), "*"),
	wire.Struct(new(LoginHintHandler), "*"),

	NewCSRFCookieDef,
	NewSessionCookieDef,
	NewClientIDCookieDef,
	NewErrorCookieDef,
	NewSignedUpCookieDef,
	wire.Struct(new(ErrorCookie), "*"),

	wire.Struct(new(CSRFMiddleware), "*"),
	wire.Struct(new(AuthEntryPointMiddleware), "*"),
	wire.Struct(new(SessionMiddleware), "*"),
	wire.Bind(new(SessionMiddlewareStore), new(*SessionStoreRedis)),
	wire.Struct(new(UILocalesMiddleware), "*"),
	wire.Struct(new(WeChatRedirectURIMiddleware), "*"),
	wire.Struct(new(ClientIDMiddleware), "*"),

	NewPublicOriginMiddlewareLogger,
	wire.Struct(new(PublicOriginMiddleware), "*"),

	NewServiceLogger,
	wire.Struct(new(SessionStoreRedis), "*"),
	wire.Bind(new(SessionStore), new(*SessionStoreRedis)),
	wire.Struct(new(Service2), "*"),
	wire.Bind(new(AuthenticateURLPageService), new(*Service2)),
	wire.Bind(new(LoginHintPageService), new(*Service2)),

	wire.Struct(new(WechatURLProvider), "*"),
)

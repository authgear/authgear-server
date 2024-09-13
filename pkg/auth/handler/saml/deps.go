package saml

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	NewLoginHandlerLogger,
	NewLoginFinishHandlerLogger,
	NewLogoutHandlerLogger,
	wire.Struct(new(LoginResultHandler), "*"),
	wire.Struct(new(MetadataHandler), "*"),
	wire.Struct(new(LoginHandler), "*"),
	wire.Struct(new(LoginFinishHandler), "*"),
	wire.Struct(new(LogoutHandler), "*"),
)

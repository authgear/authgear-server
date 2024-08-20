package saml

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	NewLoginHandlerLogger,
	NewLoginFinishHandlerLogger,
	wire.Struct(new(MetadataHandler), "*"),
	wire.Struct(new(LoginHandler), "*"),
	wire.Struct(new(LoginFinishHandler), "*"),
)

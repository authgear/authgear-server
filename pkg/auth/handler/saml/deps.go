package saml

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	NewLoginHandlerLogger,
	wire.Struct(new(MetadataHandler), "*"),
	wire.Struct(new(LoginHandler), "*"),
)

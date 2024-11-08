package whatsapp

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	NewServiceLogger,
	NewHTTPClient,
	NewWhatsappOnPremisesClient,
	wire.Struct(new(TokenStore), "*"),
	wire.Struct(new(Service), "*"),
)

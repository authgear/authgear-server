package whatsapp

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	NewServiceLogger,
	NewWhatsappOnPremisesClient,
	wire.Struct(new(TokenStore), "*"),
	wire.Struct(new(Service), "*"),
)

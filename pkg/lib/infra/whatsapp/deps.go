package whatsapp

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	NewWhatsappOnPremisesClient,
	wire.Struct(new(TokenStore), "*"),
	wire.Struct(new(Client), "*"),
)

package whatsapp

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	NewClientLogger,
	NewWhatsappOnPremisesClient,
	wire.Struct(new(TokenStore), "*"),
	wire.Struct(new(Client), "*"),
)

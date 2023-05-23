package whatsapp

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	NewWhatsappOnPremisesClient,
)

package wellknown

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(AppleAppSiteAssociationHandler), "*"),
	wire.Struct(new(AssetLinksHandler), "*"),
)

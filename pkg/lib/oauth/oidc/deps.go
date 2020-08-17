package oidc

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(MetadataProvider), "*"),
	wire.Struct(new(IDTokenIssuer), "*"),
)

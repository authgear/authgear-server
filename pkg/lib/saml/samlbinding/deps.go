package samlbinding

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(SAMLBindingHTTPPostWriter), "*"),
	wire.Struct(new(SAMLBindingHTTPRedirectWriter), "*"),
)

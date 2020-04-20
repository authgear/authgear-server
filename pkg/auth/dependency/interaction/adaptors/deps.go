package adaptors

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(IdentityAdaptor), "*"),
	wire.Struct(new(AuthenticatorAdaptor), "*"),
)

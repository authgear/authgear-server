package loader

import "github.com/google/wire"

var DependencySet = wire.NewSet(
	wire.Struct(new(UserLoader), "*"),
	wire.Struct(new(IdentityLoader), "*"),
	wire.Struct(new(AuthenticatorLoader), "*"),
)

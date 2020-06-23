package flows

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(WebAppFlow), "*"),
	wire.Struct(new(AnonymousFlow), "*"),
	wire.Struct(new(PasswordFlow), "*"),
	wire.Struct(new(UserController), "*"),
)

package forgotpassword

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(Service), "*"),
	wire.Struct(new(Sender), "*"),
	NewLogger,
)

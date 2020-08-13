package clock

import "github.com/google/wire"

var DependencySet = wire.NewSet(
	wire.InterfaceValue(new(Clock), NewSystemClock()),
)

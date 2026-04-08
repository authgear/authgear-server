package usage

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(GlobalDBStore), "*"),
	wire.Struct(new(CountCollector), "*"),
	wire.Struct(new(UsageAlertEmailServiceImpl), "*"),
	wire.Bind(new(UsageAlertEmailService), new(*UsageAlertEmailServiceImpl)),
	wire.Struct(new(Limiter), "*"),
)

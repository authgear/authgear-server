package analytic

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(GlobalDBStore), "*"),
	wire.Struct(new(AppDBStore), "*"),
	wire.Struct(new(AuditDBReadStore), "*"),
	wire.Struct(new(AuditDBWriteStore), "*"),
	wire.Struct(new(UserWeeklyReport), "*"),
	wire.Struct(new(ProjectHourlyReport), "*"),
	wire.Struct(new(ProjectWeeklyReport), "*"),
	wire.Struct(new(ProjectMonthlyReport), "*"),
	wire.Struct(new(CountCollector), "*"),
	wire.Struct(new(ChartService), "*"),
	wire.Struct(new(Service), "*"),
	wire.Struct(new(PosthogIntegration), "*"),
	NewPosthogHTTPClient,
	NewPosthogLogger,
)

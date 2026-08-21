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
	wire.Struct(new(PosthogService), "*"),
	NewPosthogHTTPClient,
)

// FirstAuthSinkDependencySet provides the real-time application.first_auth
// event sink for server binaries that dispatch auth events. Unlike
// DependencySet, it pulls in no batch/report machinery.
var FirstAuthSinkDependencySet = wire.NewSet(
	NewPosthogHTTPClient,
	NewPosthogCredentials,
	wire.Struct(new(PosthogService), "*"),
	wire.Struct(new(FirstAuthSink), "*"),
)

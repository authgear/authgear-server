package analytic

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	NewStoreRedisLogger,
	wire.Struct(new(GlobalDBStore), "*"),
	wire.Struct(new(AppDBStore), "*"),
	wire.Struct(new(AuditDBReadStore), "*"),
	wire.Struct(new(AuditDBWriteStore), "*"),
	wire.Struct(new(UserWeeklyReport), "*"),
	wire.Struct(new(ProjectWeeklyReport), "*"),
	wire.Struct(new(Service), "*"),
	wire.Struct(new(ReadStoreRedis), "*"),
	wire.Struct(new(WriteStoreRedis), "*"),
	wire.Bind(new(CounterStore), new(*WriteStoreRedis)),
	wire.Struct(new(CountCollector), "*"),
)

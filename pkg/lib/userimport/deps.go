package userimport

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(UserImportService), "*"),
	wire.Struct(new(StoreRedis), "*"),
	wire.Struct(new(JobManager), "*"),
	wire.Bind(new(Store), new(*StoreRedis)),
	NewLogger,
)

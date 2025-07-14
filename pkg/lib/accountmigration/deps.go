package accountmigration

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	NewHookHTTPClient,
	NewHookDenoClient,
	wire.Struct(new(Service), "*"),
	wire.Struct(new(AccountMigrationWebHook), "*"),
	wire.Struct(new(AccountMigrationDenoHook), "*"),
)

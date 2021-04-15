package db

import (
	globaldb "github.com/authgear/authgear-server/pkg/lib/infra/db/global"
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	globaldb.NewHandle,
	globaldb.NewSQLBuilder,
	globaldb.NewSQLExecutor,
)

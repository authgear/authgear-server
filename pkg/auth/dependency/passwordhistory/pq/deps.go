package pq

import (
	"github.com/google/wire"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/passwordhistory"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/time"
)

func ProvidePasswordHistoryStore(
	tp time.Provider,
	sqlb db.SQLBuilder,
	sqle db.SQLExecutor,
) passwordhistory.Store {
	return NewPasswordHistoryStore(tp, sqlb, sqle)
}

var DependencySet = wire.NewSet(ProvidePasswordHistoryStore)

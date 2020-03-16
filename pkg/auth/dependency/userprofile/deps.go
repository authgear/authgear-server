package userprofile

import (
	"github.com/google/wire"

	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/time"
)

func ProvideStore(tp time.Provider, sqlb db.SQLBuilder, sqle db.SQLExecutor) Store {
	return NewUserProfileStore(tp, sqlb, sqle)
}

var DependencySet = wire.NewSet(
	ProvideStore,
)

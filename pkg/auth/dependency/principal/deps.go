package principal

import (
	"github.com/google/wire"
	"github.com/skygeario/skygear-server/pkg/core/db"
)

func ProvideIdentityProvider(
	sqlb db.SQLBuilder,
	sqle db.SQLExecutor,
	ps []Provider,
) IdentityProvider {
	return NewIdentityProvider(sqlb, sqle, ps...)
}

var DependencySet = wire.NewSet(ProvideIdentityProvider)

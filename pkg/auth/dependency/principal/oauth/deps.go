package oauth

import (
	"github.com/google/wire"
	"github.com/skygeario/skygear-server/pkg/core/db"
)

func ProvideOAuthProvider(
	sqlb db.SQLBuilder,
	sqle db.SQLExecutor,
) Provider {
	return NewProvider(sqlb, sqle)
}

var DependencySet = wire.NewSet(ProvideOAuthProvider)

package anonymous

import (
	"github.com/google/wire"
	"github.com/skygeario/skygear-server/pkg/db"
)

func ProvideProvider(
	sqlb db.SQLBuilder,
	sqle db.SQLExecutor,
) *Provider {
	return &Provider{
		Store: &Store{SQLBuilder: sqlb, SQLExecutor: sqle},
	}
}

var DependencySet = wire.NewSet(ProvideProvider)

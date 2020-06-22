package oauth

import (
	"github.com/google/wire"

	"github.com/skygeario/skygear-server/pkg/clock"
	"github.com/skygeario/skygear-server/pkg/db"
)

func ProvideProvider(
	sqlb db.SQLBuilder,
	sqle db.SQLExecutor,
	t clock.Clock,
) *Provider {
	return &Provider{
		Store: &Store{SQLBuilder: sqlb, SQLExecutor: sqle},
		Clock: t,
	}
}

var DependencySet = wire.NewSet(ProvideProvider)

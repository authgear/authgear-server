package oauth

import (
	"github.com/google/wire"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/time"
)

func ProvideProvider(
	sqlb db.SQLBuilder,
	sqle db.SQLExecutor,
	t time.Provider,
) *Provider {
	return &Provider{
		Store: &Store{SQLBuilder: sqlb, SQLExecutor: sqle},
		Time:  t,
	}
}

var DependencySet = wire.NewSet(ProvideProvider)

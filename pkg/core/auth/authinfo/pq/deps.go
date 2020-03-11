package pq

import (
	"github.com/google/wire"

	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/db"
)

func ProvideStore(f db.SQLBuilderFactory, exec db.SQLExecutor) authinfo.Store {
	return NewAuthInfoStore(f("core"), exec)
}

var DependencySet = wire.NewSet(
	ProvideStore,
)

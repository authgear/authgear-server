package siteadmin

import (
	"time"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
)

// newSiteadminGlobalHandle constructs a *globaldb.Handle for the siteadmin server
// using ConnectionPurposeSiteadminGlobal so that siteadmin gets its own connection
// pool, isolated from the global DB pools of other components (portal, APIs, etc.).
func newSiteadminGlobalHandle(
	pool *db.Pool,
	credentials *config.GlobalDatabaseCredentialsEnvironmentConfig,
	cfg *config.DatabaseEnvironmentConfig,
) *globaldb.Handle {
	info := db.ConnectionInfo{
		Purpose:     db.ConnectionPurposeSiteadminGlobal,
		DatabaseURL: credentials.DatabaseURL,
	}
	opts := db.ConnectionOptions{
		MaxOpenConnection:     cfg.MaxOpenConn,
		MaxIdleConnection:     cfg.MaxIdleConn,
		MaxConnectionLifetime: time.Second * time.Duration(cfg.ConnMaxLifetimeSeconds),
		IdleConnectionTimeout: time.Second * time.Duration(cfg.ConnMaxIdleTimeSeconds),
	}
	return &globaldb.Handle{
		HookHandle: db.NewHookHandle(pool, info, opts),
	}
}

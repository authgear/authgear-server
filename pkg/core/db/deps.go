package db

import (
	"context"

	"github.com/google/wire"

	"github.com/skygeario/skygear-server/pkg/auth/config"
	oldconfig "github.com/skygeario/skygear-server/pkg/core/config"
)

func ProvideContext(ctx context.Context, pool *Pool, c *config.DatabaseCredentials) Context {
	return NewContext(ctx, pool, c)
}

// FIXME: delete this
func ProvideContextOLD(ctx context.Context, pool *Pool, c *oldconfig.TenantConfiguration) Context {
	return NewContext(ctx, pool, &config.DatabaseCredentials{
		DatabaseURL:    c.DatabaseConfig.DatabaseURL,
		DatabaseSchema: c.DatabaseConfig.DatabaseSchema,
	})
}

func ProvideSQLBuilder(c *config.DatabaseCredentials, id config.AppID) SQLBuilder {
	return NewSQLBuilder("auth", c.DatabaseSchema, string(id))
}

// FIXME: delete this
func ProvideSQLBuilderOLD(c *oldconfig.TenantConfiguration) SQLBuilder {
	return NewSQLBuilder("auth", c.DatabaseConfig.DatabaseSchema, c.AppID)
}

func ProvideSQLExecutor(ctx Context) SQLExecutor {
	return NewSQLExecutor(ctx)
}

var DependencySet = wire.NewSet(
	ProvideContext,
	ProvideSQLBuilder,
	ProvideSQLExecutor,
)

var OldDependencySet = wire.NewSet(
	ProvideContextOLD,
	ProvideSQLBuilderOLD,
	ProvideSQLExecutor,
)

package db

import (
	"context"

	"github.com/google/wire"

	"github.com/skygeario/skygear-server/pkg/core/config"
)

type Namespace string

type SQLBuilderFactory func(ns Namespace) SQLBuilder

func ProvideSQLBuilderFactory(c *config.TenantConfiguration) SQLBuilderFactory {
	return func(ns Namespace) SQLBuilder {
		return NewSQLBuilder(string(ns), c.DatabaseConfig.DatabaseSchema, c.AppID)
	}
}

func ProvideSQLExecutor(ctx context.Context, c *config.TenantConfiguration) SQLExecutor {
	return NewSQLExecutor(ctx, NewContextWithContext(ctx, *c))
}

func ProvideTxContext(ctx context.Context, c *config.TenantConfiguration) TxContext {
	return NewTxContextWithContext(ctx, *c)
}

var DependencySet = wire.NewSet(
	ProvideSQLBuilderFactory,
	ProvideSQLExecutor,
	ProvideTxContext,
)

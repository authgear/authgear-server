package db

import (
	"context"
	"database/sql"

	"github.com/skygeario/skygear-server/pkg/core/config"
)

// DBProvider is providing a postgres DB connection instance from
// connection poll that is ready to use by the tenant unware gear business
// code.
type DBProvider struct {
	namespace string
}

func NewDBProvider(gear string) *DBProvider {
	return &DBProvider{
		namespace: gear,
	}
}

func (p DBProvider) Provide(ctx context.Context, tConfig config.TenantConfiguration) *sql.Conn {
	db, err := sql.Open("postgres", tConfig.DBConnectionStr)
	if err != nil {
		panic(err)
	}
	conn, err := db.Conn(ctx)
	if err != nil {
		panic(err)
	}
	return conn
}

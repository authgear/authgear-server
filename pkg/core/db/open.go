package db

import (
	"github.com/jmoiron/sqlx"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

func OpenDB(tConfig config.TenantConfiguration) func() (*sqlx.DB, error) {
	return func() (*sqlx.DB, error) {
		return sqlx.Open("postgres", tConfig.DBConnectionStr)
	}
}

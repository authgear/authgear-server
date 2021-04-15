package global

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
)

type SQLBuilder struct {
	db.SQLBuilder
}

func NewSQLBuilder(c *config.DatabaseEnvironmentConfig) *SQLBuilder {
	return &SQLBuilder{
		db.NewSQLBuilder(c.DatabaseSchema, ""),
	}
}

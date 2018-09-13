
package pq

import (
	"context"

	sq "github.com/lann/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

type store struct {
	DB                     *sqlx.DB
	context                context.Context
}

func (s *store) Close() error { return s.DB.Close() }

// return the raw unquoted schema name of this app
func (s *store) schemaName() string {
	return "app_config"
}

// return the quoted table name ready to be used as identifier (in the form
// "schema"."table")
func (s *store) tableName(table string) string {
	return pq.QuoteIdentifier(s.schemaName()) + "." + pq.QuoteIdentifier(table)
}

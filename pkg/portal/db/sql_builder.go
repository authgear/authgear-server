package db

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/lib/pq"

	"github.com/authgear/authgear-server/pkg/portal/config"
)

type SQLBuilder struct {
	sq.StatementBuilderType
	Schema string
}

func NewSQLBuilder(config *config.DatabaseConfig) *SQLBuilder {
	return &SQLBuilder{
		StatementBuilderType: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
		Schema:               config.DatabaseSchema,
	}
}

func (b *SQLBuilder) FullTableName(table string) string {
	return pq.QuoteIdentifier(b.Schema) + "." + pq.QuoteIdentifier("_portal_"+table)
}

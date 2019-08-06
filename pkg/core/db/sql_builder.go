package db

import (
	"regexp"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/lib/pq"
)

var underscoreRe = regexp.MustCompile(`[.:]`)

func toLowerAndUnderscore(s string) string {
	return underscoreRe.ReplaceAllLiteralString(strings.ToLower(s), "_")
}

type SQLBuilder struct {
	sq.StatementBuilderType

	namespace string
	appName   string
}

func NewSQLBuilder(namespace string, appName string) SQLBuilder {
	return SQLBuilder{
		StatementBuilderType: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
		namespace:            namespace,
		appName:              appName,
	}
}

func (b SQLBuilder) Relname(table string) string {
	return "_" + b.namespace + "_" + table
}

func (b SQLBuilder) TableName(table string) string {
	return pq.QuoteIdentifier(b.Relname(table))
}

func (b SQLBuilder) FullTableName(table string) string {
	return pq.QuoteIdentifier(b.SchemaName()) + "." + b.TableName(table)
}

func (b SQLBuilder) SchemaName() string {
	return "app_" + toLowerAndUnderscore(b.appName)
}

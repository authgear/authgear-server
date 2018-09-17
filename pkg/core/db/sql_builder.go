package db

import (
	"regexp"
	"strings"

	sq "github.com/lann/squirrel"
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

func (b SQLBuilder) TableName(table string) string {
	tableName := "_" + b.namespace + "_" + table
	return pq.QuoteIdentifier(b.schemaName()) + "." + pq.QuoteIdentifier(tableName)
}

func (b SQLBuilder) schemaName() string {
	return "app_" + toLowerAndUnderscore(b.appName)
}

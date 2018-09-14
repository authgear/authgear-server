package sql

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

type Builder struct {
	sq.StatementBuilderType

	namespace string
	appName   string
}

func NewBuilder(namespace string, appName string) Builder {
	return Builder{
		StatementBuilderType: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
		namespace:            namespace,
		appName:              appName,
	}
}

func (b Builder) TableName(table string) string {
	return pq.QuoteIdentifier(b.schemaName()) + "." + pq.QuoteIdentifier(table)
}

func (b Builder) schemaName() string {
	return "app_" + toLowerAndUnderscore(b.appName)
}

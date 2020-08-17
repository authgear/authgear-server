package db

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/lib/pq"
)

type SQLBuilder struct {
	builder sq.StatementBuilderType

	namespace string
	schema    string
	appID     string
	forTenant bool
}

func newSQLBuilder() sq.StatementBuilderType {
	return sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
}

func NewSQLBuilder(namespace string, schema string, appID string) SQLBuilder {
	return SQLBuilder{
		builder:   newSQLBuilder(),
		namespace: namespace,
		schema:    schema,
		appID:     appID,
	}
}

func (b SQLBuilder) FullTableName(table string) string {
	return pq.QuoteIdentifier(b.schema) + "." + pq.QuoteIdentifier("_"+b.namespace+"_"+table)
}

func (b SQLBuilder) Tenant() SQLStatementBuilder {
	return SQLStatementBuilder{
		builder:   b.builder,
		forTenant: true,
		appID:     b.appID,
	}
}

func (b SQLBuilder) Global() SQLStatementBuilder {
	return SQLStatementBuilder{
		builder:   b.builder,
		forTenant: false,
	}
}

type SQLStatementBuilder struct {
	builder sq.StatementBuilderType

	forTenant bool
	appID     string
}

func (b SQLStatementBuilder) Select(columns ...string) SelectBuilder {
	builder := b.builder.Select(columns...)
	return SelectBuilder{
		builder:   builder,
		forTenant: b.forTenant,
		appID:     b.appID,
	}
}

func (b SQLStatementBuilder) Insert(into string) InsertBuilder {
	builder := b.builder.Insert(into)
	if b.forTenant {
		builder = builder.Columns("app_id")
	}
	return InsertBuilder{
		builder:   builder,
		forTenant: b.forTenant,
		appID:     b.appID,
	}
}

func (b SQLStatementBuilder) Update(table string) sq.UpdateBuilder {
	builder := b.builder.Update(table)
	if b.forTenant {
		builder = builder.Where("app_id = ?", b.appID)
	}
	return builder
}

func (b SQLStatementBuilder) Delete(from string) sq.DeleteBuilder {
	builder := b.builder.Delete(from)
	if b.forTenant {
		builder = builder.Where("app_id = ?", b.appID)
	}
	return builder
}

type InsertBuilder struct {
	builder   sq.InsertBuilder
	forTenant bool
	appID     string
}

// nolint: golint
func (b InsertBuilder) ToSql() (string, []interface{}, error) {
	return b.builder.ToSql()
}

func (b InsertBuilder) Columns(columns ...string) InsertBuilder {
	b.builder = b.builder.Columns(columns...)
	return b
}

func (b InsertBuilder) Values(values ...interface{}) InsertBuilder {
	if b.forTenant {
		values = append([]interface{}{b.appID}, values...)
	}
	b.builder = b.builder.Values(values...)
	return b
}

type SelectBuilder struct {
	builder   sq.SelectBuilder
	forTenant bool
	appID     string
}

// nolint: golint
func (b SelectBuilder) ToSql() (string, []interface{}, error) {
	return b.builder.ToSql()
}

func (b SelectBuilder) From(from string, alias ...string) SelectBuilder {
	if len(alias) > 0 {
		from = fmt.Sprintf("%s AS %s", from, alias[0])
		b.builder = b.builder.From(from)
		if b.forTenant {
			b.builder = b.builder.Where(alias[0]+".app_id = ?", b.appID)
		}
	} else {
		b.builder = b.builder.From(from)
		if b.forTenant {
			b.builder = b.builder.Where("app_id = ?", b.appID)
		}
	}
	return b
}

func (b SelectBuilder) Join(from string, alias string, pred string, args ...interface{}) SelectBuilder {
	join := fmt.Sprintf("%s AS %s ON %s", from, alias, pred)
	b.builder = b.builder.Join(join, args...)
	if b.forTenant {
		b.builder = b.builder.Where(alias+".app_id = ?", b.appID)
	}
	return b
}

func (b SelectBuilder) Where(pred string, args ...interface{}) SelectBuilder {
	b.builder = b.builder.Where(pred, args...)
	return b
}

func (b SelectBuilder) OrderBy(orderBy ...string) SelectBuilder {
	b.builder = b.builder.OrderBy(orderBy...)
	return b
}

func (b SelectBuilder) Limit(limit uint64) SelectBuilder {
	b.builder = b.builder.Limit(limit)
	return b
}

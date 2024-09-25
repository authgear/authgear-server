package db

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/lib/pq"
)

type sqlBuilderSchema struct {
	Schema string
}

func (b sqlBuilderSchema) TableName(table string) string {
	return pq.QuoteIdentifier(b.Schema) + "." + pq.QuoteIdentifier(table)
}

func newStatementBuilderType() sq.StatementBuilderType {
	return sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
}

type SQLBuilder struct {
	sq.StatementBuilderType
	sqlBuilderSchema
}

func NewSQLBuilder(schema string) SQLBuilder {
	return SQLBuilder{
		StatementBuilderType: newStatementBuilderType(),
		sqlBuilderSchema: sqlBuilderSchema{
			Schema: schema,
		},
	}
}

type SQLBuilderApp struct {
	sqlBuilderSchema
	builder sq.StatementBuilderType
	appID   string
}

func NewSQLBuilderApp(schema string, appID string) SQLBuilderApp {
	return SQLBuilderApp{
		builder: newStatementBuilderType(),
		sqlBuilderSchema: sqlBuilderSchema{
			Schema: schema,
		},
		appID: appID,
	}
}

func (b SQLBuilderApp) Select(columns ...string) SelectBuilder {
	builder := b.builder.Select(columns...)
	return SelectBuilder{
		builder: builder,
		appID:   b.appID,
	}
}

func (b SQLBuilderApp) Insert(into string) InsertBuilder {
	builder := b.builder.Insert(into)
	builder = builder.Columns("app_id")
	return InsertBuilder{
		builder: builder,
		appID:   b.appID,
	}
}

func (b SQLBuilderApp) Update(table string) sq.UpdateBuilder {
	builder := b.builder.Update(table)
	builder = builder.Where("app_id = ?", b.appID)
	return builder
}

func (b SQLBuilderApp) Delete(from string) sq.DeleteBuilder {
	builder := b.builder.Delete(from)
	builder = builder.Where("app_id = ?", b.appID)
	return builder
}

type InsertBuilder struct {
	builder sq.InsertBuilder
	appID   string
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
	values = append([]interface{}{b.appID}, values...)
	b.builder = b.builder.Values(values...)
	return b
}

func (b InsertBuilder) Suffix(sql string, args ...interface{}) InsertBuilder {
	b.builder = b.builder.Suffix(sql, args...)
	return b
}

type SelectBuilder struct {
	builder sq.SelectBuilder
	appID   string
}

// nolint: golint
func (b SelectBuilder) ToSql() (string, []interface{}, error) {
	return b.builder.ToSql()
}

func (b SelectBuilder) From(from string, alias ...string) SelectBuilder {
	if len(alias) > 0 {
		from = fmt.Sprintf("%s AS %s", from, alias[0])
		b.builder = b.builder.From(from)
		b.builder = b.builder.Where(alias[0]+".app_id = ?", b.appID)
	} else {
		b.builder = b.builder.From(from)
		b.builder = b.builder.Where("app_id = ?", b.appID)
	}
	return b
}

func (b SelectBuilder) Join(from string, alias string, pred string, args ...interface{}) SelectBuilder {
	join := fmt.Sprintf("%s AS %s ON %s", from, alias, pred)
	b.builder = b.builder.Join(join, args...)
	b.builder = b.builder.Where(alias+".app_id = ?", b.appID)
	return b
}

func (b SelectBuilder) LeftJoin(from string, alias string, pred string, args ...interface{}) SelectBuilder {
	join := fmt.Sprintf("%s AS %s ON %s", from, alias, pred)
	b.builder = b.builder.LeftJoin(join, args...)
	b.builder = b.builder.Where(alias+".app_id = ?", b.appID)
	return b
}

func (b SelectBuilder) Where(pred interface{}, args ...interface{}) SelectBuilder {
	b.builder = b.builder.Where(pred, args...)
	return b
}

func (b SelectBuilder) PrefixExpr(expr sq.Sqlizer) SelectBuilder {
	b.builder = b.builder.PrefixExpr(expr)
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

func (b SelectBuilder) Offset(offset uint64) SelectBuilder {
	b.builder = b.builder.Offset(offset)
	return b
}

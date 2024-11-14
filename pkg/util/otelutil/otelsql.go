package otelutil

import (
	"context"
	"database/sql"
	"database/sql/driver"

	"github.com/XSAM/otelsql"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/semconv/v1.27.0"
)

func otelsqlInstrumentAttributesGetter(ctx context.Context, method otelsql.Method, query string, args []driver.NamedValue) []attribute.KeyValue {
	res := GetResource(ctx)
	attrs := ExtractAttributesFromResource(res)
	return attrs
}

// OTelSQLOpen is database/sql.Open, with instrumentation.
func OTelSQLOpenPostgres(connectionURL string) (*sql.DB, error) {
	options := []otelsql.Option{
		otelsql.WithAttributes(semconv.DBSystemPostgreSQL),
		otelsql.WithInstrumentAttributesGetter(otelsqlInstrumentAttributesGetter),
	}

	db, err := otelsql.Open("postgres", connectionURL, options...)
	if err != nil {
		return nil, err
	}

	err = otelsql.RegisterDBStatsMetrics(db, options...)
	if err != nil {
		return nil, err
	}

	return db, nil
}

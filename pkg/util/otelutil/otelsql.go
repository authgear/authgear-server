package otelutil

import (
	"database/sql"

	"github.com/XSAM/otelsql"

	"go.opentelemetry.io/otel/semconv/v1.27.0"
)

// OTelSQLOpen is database/sql.Open, with instrumentation.
func OTelSQLOpenPostgres(connectionURL string) (*sql.DB, error) {
	options := []otelsql.Option{
		otelsql.WithAttributes(semconv.DBSystemPostgreSQL),
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

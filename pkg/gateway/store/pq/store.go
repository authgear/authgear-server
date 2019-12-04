package pq

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/gateway/store"
)

// NewGatewayStore create new gateway store by db connection url
func NewGatewayStore(ctx context.Context, pool db.Pool, connString string) (*Store, error) {
	db, err := pool.OpenURL(connString)
	if err != nil {
		return nil, err
	}

	return &Store{
		DB:      db,
		context: ctx,
	}, nil
}

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

type Store struct {
	DB      *sqlx.DB
	context context.Context
}

func (s *Store) Close() error { return s.DB.Close() }

// return the raw unquoted schema name of this app
func (s *Store) schemaName() string {
	return "app_config"
}

// return the quoted table name ready to be used as identifier (in the form
// "schema"."table")
func (s *Store) tableName(table string) string {
	return pq.QuoteIdentifier(s.schemaName()) + "." + pq.QuoteIdentifier(table)
}

// this ensures that our structure conform to certain interfaces.
var (
	_ store.GatewayStore = &Store{}
)

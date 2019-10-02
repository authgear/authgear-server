package pq

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"

	"github.com/skygeario/skygear-server/pkg/core/logging"
	"github.com/skygeario/skygear-server/pkg/gateway/store"
)

// NewGatewayStore create new gateway store by db connection url
func NewGatewayStore(ctx context.Context, connString string, loggerFactory logging.Factory) (*Store, error) {
	s, err := Connect(ctx, connString)
	if err != nil {
		return nil, err
	}
	s.logger = loggerFactory.NewLogger("gateway-store")
	return s, nil
}

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

type Store struct {
	DB      *sqlx.DB
	context context.Context
	logger  *logrus.Entry
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

// Connect returns a new connection to postgresql implementation
func Connect(ctx context.Context, connString string) (*Store, error) {
	db, err := sqlx.Connect("postgres", connString)
	if err != nil {
		return nil, fmt.Errorf("failed to open connection: %s", err)
	}

	return &Store{
		DB:      db,
		context: ctx,
	}, nil
}

// this ensures that our structure conform to certain interfaces.
var (
	_ store.GatewayStore = &Store{}
)

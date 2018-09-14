package pq

import (
	"fmt"
	"context"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// Connect returns a new connection to postgresql implementation
func Connect(ctx context.Context, connString string) (*store, error) {
	db, err := sqlx.Connect("postgres", connString)
	if err != nil {
		return nil, fmt.Errorf("failed to open connection: %s", err)
	}

	return &store{
		DB:                     db,
		context:                ctx,
	}, nil
}

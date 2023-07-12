package internal

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	sq "github.com/Masterminds/squirrel"
	"github.com/lib/pq"
)

func openDB(dbURL string, dbSchema string) *sql.DB {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("failed to open database: %s", err)
	}

	_, err = db.Exec(fmt.Sprintf("SET search_path TO %s", pq.QuoteIdentifier(dbSchema)))
	if err != nil {
		log.Fatalf("failed to set search_path: %s", err)
	}

	return db
}

func newSQLBuilder() sq.StatementBuilderType {
	return sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
}

func WithTx(ctx context.Context, db *sql.DB, do func(tx *sql.Tx) error) (err error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		} else if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	err = do(tx)
	return
}

type Queryer interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

type ConfigSource struct {
	ID    string
	AppID string
	Data  map[string]string
}

func selectConfigSources(ctx context.Context, queryer Queryer, appID []string) ([]*ConfigSource, error) {
	builder := newSQLBuilder().
		Select(
			"id",
			"app_id",
			"data",
		).
		From(pq.QuoteIdentifier("_portal_config_source"))
	if len(appID) > 0 {
		builder = builder.Where("app_id = ANY (?)", pq.Array(appID))
	}

	q, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := queryer.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}

	var result []*ConfigSource
	for rows.Next() {
		i := ConfigSource{}
		var data []byte

		if err := rows.Scan(
			&i.ID,
			&i.AppID,
			&data,
		); err != nil {
			return nil, err
		}
		if err = json.Unmarshal(data, &i.Data); err != nil {
			return nil, err
		}
		result = append(result, &i)
	}

	return result, nil
}

func updateConfigSource(ctx context.Context, tx *sql.Tx, source *ConfigSource) error {
	dataBytes, err := json.Marshal(source.Data)
	if err != nil {
		return err
	}

	builder := newSQLBuilder().
		Update(pq.QuoteIdentifier("_portal_config_source")).
		Set("data", dataBytes).
		Where("id = ?", source.ID)

	q, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	result, err := tx.ExecContext(ctx, q, args...)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("config sources not found, id: %s", source.ID)
	} else if rowsAffected > 1 {
		panic(fmt.Errorf("config sources want 1 row updated, got %v, id: %s", rowsAffected, source.ID))
	}

	return nil
}

package internal

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"reflect"

	"github.com/lib/pq"
)

type Queryer interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

type MigrateResourcesOptions struct {
	DatabaseURL            string
	DatabaseSchema         string
	DryRun                 *bool
	UpdateConfigSourceFunc func(appID string, configSourceData map[string]string) error
}

type configSource struct {
	ID      string
	AppID   string
	Data    map[string]string
	Updated bool
}

func MigrateResources(opt *MigrateResourcesOptions) {
	db := openDB(opt.DatabaseURL, opt.DatabaseSchema)

	ctx := context.Background()
	configSourceList, err := selectAllConfigSources(ctx, db)
	if err != nil {
		log.Fatalf("failed to connect db: %s", err)
	}

	for _, c := range configSourceList {
		original := make(map[string]string)
		for k, v := range c.Data {
			original[k] = v
		}

		if err := opt.UpdateConfigSourceFunc(c.AppID, c.Data); err != nil {
			log.Fatalf("failed to convert resources: %s, %s", c.AppID, err)
		}

		c.Updated = !reflect.DeepEqual(original, c.Data)
		log.Printf("converting resources app_id: %s, updated: %t", c.AppID, c.Updated)

	}

	// dryRun default is true
	dryRun := true
	if opt.DryRun != nil {
		dryRun = *opt.DryRun
	}
	if dryRun {
		count := 0
		for _, c := range configSourceList {
			if c.Updated {
				log.Printf("dry run: resources to update: appid: %s", c.AppID)
				data, err := json.MarshalIndent(c.Data, "", "  ")
				if err != nil {
					panic(err)
				}
				log.Printf("%s\n", string(data))
				count++
			}
		}
		log.Printf("dry run: number of apps to update: %d", count)
		return
	}

	// update config to db
	count := 0
	for _, c := range configSourceList {
		if c.Updated {
			count++
			err := WithTx(ctx, db, func(tx *sql.Tx) error {
				err := updateConfigSource(ctx, tx, c)
				return err
			})
			if err != nil {
				log.Fatalf("failed to update resources to db: %s, %s", c.AppID, err)
			} else {
				log.Printf("updated resources to db: %s", c.AppID)
			}
		}
	}
	log.Printf("updated apps count: %d", count)
}

func selectAllConfigSources(ctx context.Context, queryer Queryer) ([]*configSource, error) {
	builder := newSQLBuilder().
		Select(
			"id",
			"app_id",
			"data",
		).
		From(pq.QuoteIdentifier("_portal_config_source"))

	q, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := queryer.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}

	var result []*configSource
	for rows.Next() {
		i := configSource{}
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

func updateConfigSource(ctx context.Context, tx *sql.Tx, source *configSource) error {
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

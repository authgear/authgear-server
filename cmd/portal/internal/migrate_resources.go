package internal

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"reflect"
)

type MigrateResourcesOptions struct {
	DatabaseURL            string
	DatabaseSchema         string
	DryRun                 *bool
	UpdateConfigSourceFunc func(appID string, configSourceData map[string]string, DryRun bool) error
}

func MigrateResources(opt *MigrateResourcesOptions) {
	db := openDB(opt.DatabaseURL, opt.DatabaseSchema)

	ctx := context.Background()
	configSourceList, err := selectConfigSources(ctx, db, nil)
	if err != nil {
		log.Fatalf("failed to connect db: %s", err)
	}
	// dryRun default is true
	dryRun := true
	if opt.DryRun != nil {
		dryRun = *opt.DryRun
	}

	var configSourcesToUpdate []*ConfigSource
	for _, c := range configSourceList {
		original := make(map[string]string)
		for k, v := range c.Data {
			original[k] = v
		}

		if err := opt.UpdateConfigSourceFunc(c.AppID, c.Data, dryRun); err != nil {
			log.Fatalf("failed to convert resources: %s, %s", c.AppID, err)
		}

		updated := !reflect.DeepEqual(original, c.Data)
		log.Printf("converting resources app_id: %s, updated: %t", c.AppID, updated)
		if updated {
			configSourcesToUpdate = append(configSourcesToUpdate, c)
		}

		if dryRun {
			if updated {
				log.Printf("dry run: original resources appid (%s)", c.AppID)
				data, err := json.MarshalIndent(original, "", "  ")
				if err != nil {
					panic(err)
				}
				log.Printf("%s\n", string(data))

				log.Printf("dry run: updated resources appid (%s)", c.AppID)
				data, err = json.MarshalIndent(c.Data, "", "  ")
				if err != nil {
					panic(err)
				}
				log.Printf("%s\n", string(data))
			}
		}
	}

	if dryRun {
		log.Printf("dry run: number of apps to update: %d", len(configSourcesToUpdate))
		return
	}

	// update config to db
	count := 0
	for _, c := range configSourcesToUpdate {
		err := WithTx(ctx, db, func(tx *sql.Tx) error {
			err := updateConfigSource(ctx, tx, c)
			return err
		})
		if err != nil {
			log.Fatalf("failed to update resources to db: %s, %s", c.AppID, err)
		} else {
			log.Printf("updated resources to db: %s", c.AppID)
			count++
		}
	}
	log.Printf("updated apps count: %d", count)
}

package internal

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"log"
	"os"
	"os/exec"
	"reflect"
)

type MigrateResourcesOptions struct {
	DatabaseURL            string
	DatabaseSchema         string
	DryRun                 *bool
	UpdateConfigSourceFunc func(appID string, configSourceData map[string]string, DryRun bool) error
}

// nolint: gocognit
func MigrateResources(ctx context.Context, opt *MigrateResourcesOptions) {
	db := openDB(opt.DatabaseURL, opt.DatabaseSchema)

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
				appID := c.AppID
				originalAuthgearYAMLBytes, err := base64.StdEncoding.DecodeString(original["authgear.yaml"])
				if err != nil {
					panic(err)
				}

				updatedAuthgearYAMLBytes, err := base64.StdEncoding.DecodeString(c.Data["authgear.yaml"])
				if err != nil {
					panic(err)
				}

				diff, err := Diff("authgear.yaml", originalAuthgearYAMLBytes, updatedAuthgearYAMLBytes)
				if err != nil {
					panic(err)
				}

				log.Printf("diff of authgear.yaml: %v\n", appID)
				log.Printf("%v\n", diff)
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

func Diff(filename string, original []byte, updated []byte) (diff string, err error) {
	fOriginal, err := os.CreateTemp("", filename)
	if err != nil {
		return
	}
	defer os.Remove(fOriginal.Name())

	fUpdated, err := os.CreateTemp("", filename)
	if err != nil {
		return
	}
	defer os.Remove(fUpdated.Name())

	_, err = fOriginal.Write(original)
	if err != nil {
		return
	}
	err = fOriginal.Close()
	if err != nil {
		return
	}

	_, err = fUpdated.Write(updated)
	if err != nil {
		return
	}
	err = fUpdated.Close()
	if err != nil {
		return
	}

	output, err := exec.Command( // nolint:gosec
		"diff",
		"-u",
		fOriginal.Name(),
		fUpdated.Name(),
	).CombinedOutput()
	if err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			if exitError.ExitCode() == 0 || exitError.ExitCode() == 1 {
				err = nil
			}
		}
		if err != nil {
			return
		}
	}

	diff = string(output)
	return
}

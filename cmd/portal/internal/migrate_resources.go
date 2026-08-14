package internal

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"log"
	"maps"
	"os"
	"os/exec"
	"reflect"
)

type MigrateResourcesOptions struct {
	DatabaseURL    string
	DatabaseSchema string
	DryRun         *bool

	// AuthgearYAMLContains, when non-empty, narrows the config sources loaded
	// into memory to those whose decoded authgear.yaml contains this substring.
	//
	// This exists purely for speed. A config source's data holds every resource
	// the app has customised -- templates, translations, base64 images -- so
	// loading all of them costs far more than the authgear.yaml the migration
	// actually looks at. Filtering in the database avoids shipping (and
	// allocating) the payloads of apps the migration would skip anyway.
	//
	// It MUST be a necessary condition for UpdateConfigSourceFunc to make a
	// change, otherwise apps are silently skipped.
	AuthgearYAMLContains string

	UpdateConfigSourceFunc func(ctx context.Context, appID string, configSourceData map[string]string, DryRun bool) error
}

// nolint: gocognit
func MigrateResources(ctx context.Context, opt *MigrateResourcesOptions) {
	db := openDB(opt.DatabaseURL, opt.DatabaseSchema)

	var appIDs []string
	if opt.AuthgearYAMLContains != "" {
		var err error
		appIDs, err = selectAppIDsByAuthgearYAMLSubstring(ctx, db, opt.AuthgearYAMLContains)
		if err != nil {
			log.Fatalf("failed to select candidate app ids: %s", err)
		}
		log.Printf("narrowed to %d apps whose authgear.yaml contains %q", len(appIDs), opt.AuthgearYAMLContains)
		if len(appIDs) == 0 {
			// selectConfigSources treats an empty slice as "no filter", so stop
			// here rather than loading every config source.
			log.Printf("nothing to do")
			return
		}
	}

	configSourceList, err := selectConfigSources(ctx, db, appIDs)
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
		maps.Copy(original, c.Data)

		if err := opt.UpdateConfigSourceFunc(ctx, c.AppID, c.Data, dryRun); err != nil {
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

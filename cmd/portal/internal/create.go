package internal

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/lib/pq"
	"github.com/spf13/afero"
	"sigs.k8s.io/yaml"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/filepathutil"
	"github.com/authgear/authgear-server/pkg/util/resource"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

type CreateOptions struct {
	DatabaseURL    string
	DatabaseSchema string
	ResourceDir    string
}

func Create(ctx context.Context, opt *CreateOptions) error {
	// construct config source
	data, err := constructConfigSourceData(opt.ResourceDir)
	if err != nil {
		return fmt.Errorf("invalid resource directory: %w", err)
	}

	err = validateConfigSource(data)
	if err != nil {
		return fmt.Errorf("invalid resource directory: %w", err)
	}

	appID, err := parseAppID(data["authgear.yaml"])
	if err != nil {
		return fmt.Errorf("failed to parse app id: %w", err)
	}

	// start store domains and config source to db
	db := openDB(opt.DatabaseURL, opt.DatabaseSchema)

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	if err := createConfigSource(ctx, tx, appID, data); err != nil {
		return fmt.Errorf("failed to create config source record: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// create config source record in db
func createConfigSource(ctx context.Context, tx *sql.Tx, appID string, data map[string]string) error {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return err
	}

	builder := newSQLBuilder().
		Insert(pq.QuoteIdentifier("_portal_config_source")).
		Columns(
			"id",
			"app_id",
			"data",
			"plan_name",
			"created_at",
			"updated_at",
		).
		Values(
			uuid.New(),
			appID,
			dataJSON,
			"",
			time.Now().UTC(),
			time.Now().UTC(),
		)

	q, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, q, args...)
	if err != nil {
		return err
	}

	return nil
}

func constructConfigSourceData(resourceDir string) (map[string]string, error) {
	fs := afero.NewBasePathFs(afero.NewOsFs(), resourceDir)
	appFs := &resource.LeveledAferoFs{Fs: fs, FsLevel: resource.FsLevelApp}

	locations, err := resource.EnumerateAllLocations(appFs)
	if err != nil {
		return nil, err
	}

	manager := resource.NewManager(resource.DefaultRegistry, []resource.Fs{appFs})
	var matches []resource.Location
	for _, l := range locations {
		for _, desc := range manager.Registry.Descriptors {
			if _, ok := desc.MatchResource(l.Path); ok {
				matches = append(matches, l)
				break
			}
		}
	}

	// Read the files to construct config source data
	dbData := make(map[string]string)
	for _, l := range matches {
		path := l.Path
		f, err := fs.Open(l.Path)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		data, err := ioutil.ReadAll(f)
		if err != nil {
			return nil, err
		}

		str := base64.StdEncoding.EncodeToString(data)
		dbData[filepathutil.EscapePath(path)] = str
	}

	return dbData, nil
}

func validateConfigSource(data map[string]string) error {
	// make sure the resources folder has required files
	if _, ok := data["authgear.yaml"]; !ok {
		return fmt.Errorf("missing authgear.yaml")
	}
	if _, ok := data["authgear.secrets.yaml"]; !ok {
		return fmt.Errorf("missing authgear.secrets.yaml")
	}

	return nil
}

func parseAppID(base64EncodedAuthgearYAML string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(base64EncodedAuthgearYAML)
	if err != nil {
		return "", err
	}

	cfg := config.AppConfig{}
	if err := yaml.Unmarshal(decoded, &cfg); err != nil {
		return "", fmt.Errorf("malformed authgear.yaml: %w", err)
	}

	if cfg.ID == "" {
		return "", fmt.Errorf("missing app id")
	}

	return string(cfg.ID), nil
}

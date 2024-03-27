package e2e

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"io/fs"

	"io"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/infra/task"
	"github.com/authgear/authgear-server/pkg/util/filepathutil"
	"github.com/authgear/authgear-server/pkg/util/resource"
	"github.com/authgear/authgear-server/pkg/util/uuid"
	"github.com/lib/pq"
	cp "github.com/otiai10/copy"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
)

type End2End struct {
	Context context.Context
}

type NoopTaskQueue struct{}

func (q NoopTaskQueue) Enqueue(param task.Param) {
}

func (c *End2End) CreateConfigSource(appID string, baseConfigSourceDir string) error {
	cfg, err := LoadConfigFromEnv()
	if err != nil {
		return err
	}

	// create app-specific temp config source
	configSourceDir, err := c.createTempConfigSource(appID, baseConfigSourceDir)
	if err != nil {
		return err
	}

	// construct config source
	data, err := constructConfigSourceData(configSourceDir)
	if err != nil {
		return fmt.Errorf("invalid resource directory: %w", err)
	}

	err = validateConfigSource(data)
	if err != nil {
		return fmt.Errorf("invalid resource directory: %w", err)
	}

	// start store domains and config source to db
	db := openDB(cfg.GlobalDatabase.DatabaseURL, cfg.GlobalDatabase.DatabaseSchema)

	ctx := context.Background()
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

func (c *End2End) createTempConfigSource(appID string, baseConfigSourceDir string) (string, error) {
	tempAppDir := filepath.Join(os.TempDir(), appID)

	err := cp.Copy(baseConfigSourceDir, tempAppDir)
	if err != nil {
		return "", err
	}

	authgearYAMLPath := filepath.Join(tempAppDir, configsource.AuthgearYAML)
	authgearYAML, err := os.ReadFile(authgearYAMLPath)
	if err != nil {
		return "", err
	}

	cfg := config.AppConfig{}
	err = yaml.Unmarshal(authgearYAML, &cfg)
	if err != nil {
		return "", err
	}

	cfg.ID = config.AppID(appID)
	newAuthgearYAML, err := yaml.Marshal(cfg)
	if err != nil {
		return "", err
	}

	err = os.WriteFile(authgearYAMLPath, newAuthgearYAML, fs.FileMode(0644))
	if err != nil {
		return "", err
	}

	return tempAppDir, nil
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
		data, err := io.ReadAll(io.Reader(f))
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

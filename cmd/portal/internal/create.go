package internal

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lib/pq"
	"sigs.k8s.io/yaml"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

type CreateOptions struct {
	DatabaseURL    string
	DatabaseSchema string
	ResourceDir    string
}

func Create(ctx context.Context, opt *CreateOptions) error {
	// construct config source
	data, err := pack(opt.ResourceDir)
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
	defer tx.Rollback()

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

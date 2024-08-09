package e2e

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"text/template"
)

func (c *End2End) ExecuteSQLInsertUpdate(appID string, sqlPath string) error {
	cfg, err := LoadConfigFromEnv()
	if err != nil {
		return err
	}

	db := openDB(cfg.GlobalDatabase.DatabaseURL, cfg.GlobalDatabase.DatabaseSchema)

	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	sql, err := os.ReadFile(sqlPath)
	if err != nil {
		return fmt.Errorf("failed to read SQL file: %w", err)
	}

	vars := map[string]interface{}{
		"AppID": appID,
	}

	tmpl, err := template.New("sql").Parse(string(sql))
	if err != nil {
		return fmt.Errorf("failed to parse SQL template: %w", err)
	}

	var parsedsql bytes.Buffer
	if err := tmpl.Execute(&parsedsql, vars); err != nil {
		return fmt.Errorf("failed to execute SQL template: %w", err)
	}

	if _, err := tx.ExecContext(ctx, parsedsql.String()); err != nil {
		return fmt.Errorf("failed to execute SQL: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

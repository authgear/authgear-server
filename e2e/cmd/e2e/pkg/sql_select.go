package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
)

func (c *End2End) QuerySQLSelect(appID string, rawSQL string) (jsonArrString string, err error) {
	cfg, err := LoadConfigFromEnv()
	if err != nil {
		return "", err
	}

	db := openDB(cfg.GlobalDatabase.DatabaseURL, cfg.GlobalDatabase.DatabaseSchema)

	vars := map[string]interface{}{
		"AppID": appID,
	}

	tmpl, err := ParseSQLTemplate("sql-select", rawSQL)
	if err != nil {
		return "", fmt.Errorf("failed to parse SQL template: %w", err)
	}

	var parsedsql bytes.Buffer
	if err := tmpl.Execute(&parsedsql, vars); err != nil {
		return "", fmt.Errorf("failed to execute SQL template: %w", err)
	}

	rows, err := db.Query(parsedsql.String())
	if err != nil {
		return "", fmt.Errorf("failed to execute SQL: %w", err)
	}

	parsedRows, err := ParseRows(rows)
	rows.Close()
	if err != nil {
		return "", fmt.Errorf("failed to parse SQL rows: %w", err)
	}
	jsonRows, err := json.Marshal(parsedRows)
	if err != nil {
		return "", fmt.Errorf("failed to marshal SQL rows as JSON: %w", err)
	}

	return string(jsonRows), nil
}

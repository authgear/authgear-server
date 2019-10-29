package pq

import (
	"database/sql"
	"encoding/json"

	"github.com/skygeario/skygear-server/pkg/gateway/model"
	"github.com/skygeario/skygear-server/pkg/gateway/store"
)

func (s *Store) GetLastDeploymentHooks(app model.App) (hooks *model.DeploymentHooks, err error) {
	builder := psql.Select(
		"id",
		"created_at",
		"deployment_version",
		"hooks",
	).
		From(s.tableName("deployment_hook")).
		Where("app_id = ?", app.ID).
		Where("is_last_deployment = true")

	scanner, err := s.QueryRowWith(builder)
	if err != nil {
		return
	}

	hooks = &model.DeploymentHooks{
		AppID:            app.ID,
		IsLastDeployment: true,
	}
	var hooksBytes []byte
	err = scanner.Scan(
		&hooks.ID,
		&hooks.CreatedAt,
		&hooks.DeploymentVersion,
		&hooksBytes,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			err = store.NewNotFoundError("deployment hooks")
		}
		return
	}

	err = json.Unmarshal(hooksBytes, &hooks.Hooks)
	if err != nil {
		return
	}
	return
}

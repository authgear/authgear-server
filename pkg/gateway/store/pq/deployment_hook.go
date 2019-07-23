package pq

import (
	"database/sql"
	"encoding/json"

	"github.com/skygeario/skygear-server/pkg/gateway/model"
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

	scanner := s.QueryRowWith(builder)

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
			// in case no rows exist: deployment has no hooks
			// ignore the error and return empty hooks
			err = nil
		}
		return
	}

	err = json.Unmarshal(hooksBytes, &hooks.Hooks)
	if err != nil {
		return
	}
	return
}

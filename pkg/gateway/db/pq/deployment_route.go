package pq

import (
	"encoding/json"

	"github.com/skygeario/skygear-server/pkg/gateway/model"
)

func (s *Store) GetLastDeploymentRoutes(app model.App) (routes []*model.DeploymentRoute, err error) {
	builder := psql.Select(
		"id",
		"created_at",
		"version",
		"path",
		"type",
		"type_config",
	).
		From(s.tableName("deployment_route")).
		Where("app_id = ?", app.ID).
		Where("is_last_deployment = true")

	rows, err := s.QueryWith(builder)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var r model.DeploymentRoute
		var typeConfigBytes []byte
		if err = rows.Scan(
			&r.ID,
			&r.CreatedAt,
			&r.Version,
			&r.Path,
			&r.Type,
			&typeConfigBytes,
		); err != nil {
			return
		}
		err = json.Unmarshal(typeConfigBytes, &r.TypeConfig)
		if err != nil {
			return
		}
		routes = append(routes, &r)
	}
	return
}

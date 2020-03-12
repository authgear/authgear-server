package pq

import (
	"bytes"
	"database/sql"

	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/errors"
	"github.com/skygeario/skygear-server/pkg/gateway/model"
	"github.com/skygeario/skygear-server/pkg/gateway/store"
)

func (s *Store) GetApp(id string) (*model.App, error) {
	builder := psql.Select("id", "name", "config_id", "plan_id", "auth_version").
		From(s.tableName("app")).
		Where("id = ?", id)
	scanner, err := s.QueryRowWith(builder)
	if err != nil {
		return nil, err
	}

	var (
		configID string
		planID   string
	)

	app := &model.App{}
	if err := scanner.Scan(
		&app.ID,
		&app.Name,
		&configID,
		&planID,
		&app.AuthVersion,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, store.NewNotFoundError("app")
		}
		return nil, err
	}

	configValue := config.TenantConfiguration{}
	if err := s.getConfigByID(configID, &configValue); err != nil {
		return nil, err
	}
	app.Config = configValue

	plan := model.Plan{}
	if err := s.getPlanByID(planID, &plan); err != nil {
		return nil, err
	}
	app.Plan = plan

	return app, nil

}

func (s *Store) getConfigByID(id string, configValue *config.TenantConfiguration) error {
	builder := psql.Select("config.config").
		From(s.tableName("config")).
		Where("config.id = ?", id)
	scanner, err := s.QueryRowWith(builder)
	if err != nil {
		return err
	}

	var json []byte
	err = scanner.Scan(&json)

	if errors.Is(err, sql.ErrNoRows) {
		return store.NewNotFoundError("config")
	}
	if err != nil {
		return err
	}

	config, err := config.NewTenantConfigurationFromJSON(bytes.NewReader(json), false)
	if err != nil {
		return errors.Newf("failed to scan tenant config: %w", err)
	}

	*configValue = *config
	return nil
}

func (s *Store) getPlanByID(id string, plan *model.Plan) error {
	builder := psql.Select(
		"id", "name", "auth_enabled", "created_at", "updated_at",
	).From(s.tableName("plan")).
		Where("plan.id = ?", id)
	scanner, err := s.QueryRowWith(builder)
	if err != nil {
		return err
	}

	err = scanner.StructScan(
		plan,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return store.NewNotFoundError("plan")
	}
	if err != nil {
		return err
	}

	return nil
}

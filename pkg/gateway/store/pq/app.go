package pq

import (
	"bytes"
	"database/sql"

	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/errors"
	"github.com/skygeario/skygear-server/pkg/gateway/model"
	"github.com/skygeario/skygear-server/pkg/gateway/store"
)

func (s *Store) GetAppByDomain(domain string, app *model.App) error {
	builder := psql.Select("app.id", "app.name", "app.config_id", "app.plan_id", "app.auth_version").
		From(s.tableName("app")).
		Join(s.tableName("domain")+" ON app.id = domain.app_id").
		Where("domain.domain = ?", domain)
	scanner, err := s.QueryRowWith(builder)
	if err != nil {
		return err
	}

	var (
		configID string
		planID   string
	)

	if err := scanner.Scan(
		&app.ID,
		&app.Name,
		&configID,
		&planID,
		&app.AuthVersion,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return store.NewNotFoundError("app")
		}
		return err
	}

	configValue := config.TenantConfiguration{}
	if err := s.getConfigByID(configID, &configValue); err != nil {
		return err
	}
	app.Config = configValue

	plan := model.Plan{}
	if err := s.getPlanByID(planID, &plan); err != nil {
		return err
	}
	app.Plan = plan

	return nil
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

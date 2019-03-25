package pq

import (
	"database/sql"
	"errors"

	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/core/logging"
	"github.com/skygeario/skygear-server/pkg/gateway/model"
)

// ErrAppNotFound is returned by Conn.GetAppByDomain when App cannot be found
// by given domain
var ErrAppNotFound = errors.New("App not found")

// ErrConfigNotFound is returned by Conn.GetAppByDomain when tenant config
// cannot be found
var ErrConfigNotFound = errors.New("Tenant config not found")

func (s *Store) GetAppByDomain(domain string, app *model.App) error {
	logger := logging.LoggerEntry("gateway")
	builder := psql.Select("app.id", "app.name", "app.config_id", "app.plan_id", "app.auth_version").
		From(s.tableName("app")).
		Join(s.tableName("domain")+" ON app.id = domain.app_id").
		Where("domain.domain = ?", domain)
	scanner := s.QueryRowWith(builder)

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
		if err == sql.ErrNoRows {
			return ErrAppNotFound
		}
		return err
	}

	configValue := newTenantConfigurationValue()
	if err := s.getConfigByID(configID, &configValue); err != nil {
		logger.WithError(err).Error("Fail to get app tenant config")
		return err
	}
	app.Config = configValue.TenantConfiguration

	plan := model.Plan{}
	if err := s.getPlanByID(planID, &plan); err != nil {
		logger.WithError(err).Error("Fail to get app plan")
		return err
	}
	app.Plan = plan

	logger.WithFields(logrus.Fields{
		"app": app,
	}).Debug("Got the app successfully")

	return nil
}

func (s *Store) getConfigByID(id string, configValue *tenantConfigurationValue) error {
	builder := psql.Select("config.config").
		From(s.tableName("config")).
		Where("config.id = ?", id)
	scanner := s.QueryRowWith(builder)

	err := scanner.Scan(
		configValue,
	)

	if err == sql.ErrNoRows {
		return ErrConfigNotFound
	}

	return err
}

func (s *Store) getPlanByID(id string, plan *model.Plan) error {
	builder := psql.Select(
		"id", "name", "auth_enabled", "created_at", "updated_at",
	).From(s.tableName("plan")).
		Where("plan.id = ?", id)
	scanner := s.QueryRowWith(builder)

	err := scanner.StructScan(
		plan,
	)

	return err
}

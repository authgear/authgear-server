package pq

import (
	"database/sql"
	"errors"

	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/core/logging"
	"github.com/skygeario/skygear-server/pkg/gateway/model"
)

// ErrCloudCodeNotFound is returned by Conn.FindLongestMatchedCloudCode when
// CloudCode cannot be found by given path
var ErrCloudCodeNotFound = errors.New("CloudCode not found")

func (s *Store) FindLongestMatchedCloudCode(path string, app model.App, cloudCode *model.CloudCode) error {
	logger := logging.LoggerEntry("gateway")
	builder := psql.Select(
		"cloud_code_route.id",
		"cloud_code_route.created_at",
		"cloud_code_route.version",
		"cloud_code_route.path",
		"cloud_code_route.target_path",
		"cloud_code_route.backend_url",
	).
		From(s.tableName("cloud_code_route")).
		Where("? LIKE path || '%'", path).
		Where("app_id = ?", app.ID).
		OrderBy("length(path) desc").
		Limit(1)
	scanner := s.QueryRowWith(builder)

	if err := scanner.Scan(
		&cloudCode.ID,
		&cloudCode.CreatedAt,
		&cloudCode.Version,
		&cloudCode.Path,
		&cloudCode.TargetPath,
		&cloudCode.BackendURL,
	); err != nil {
		if err == sql.ErrNoRows {
			return ErrCloudCodeNotFound
		}

		logger.WithFields(logrus.Fields{
			"path":  path,
			"app":   app.Name,
			"error": err,
		}).Errorf("Failed to query cloud code")

		return err
	}

	logger.WithFields(logrus.Fields{
		"path":       path,
		"cloud_code": cloudCode,
		"app":        app.Name,
	}).Debug("Cloud code matched")

	return nil
}

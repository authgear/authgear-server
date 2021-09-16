package analytic

import (
	"time"

	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
)

type GlobalDBStore struct {
	SQLBuilder  *globaldb.SQLBuilder
	SQLExecutor *globaldb.SQLExecutor
}

func (s *GlobalDBStore) GetAppOwners(rangeFrom *time.Time, rangeTo *time.Time) ([]*AppCollaborator, error) {
	builder := s.SQLBuilder.
		Select(
			"app_id",
			"user_id",
		).
		From(s.SQLBuilder.TableName("_portal_app_collaborator"))

	if rangeFrom != nil {
		builder = builder.Where("created_at >= ?", rangeFrom)
	}
	if rangeTo != nil {
		builder = builder.Where("created_at < ?", rangeTo)
	}

	builder = builder.
		Where("role = ?", "owner")

	rows, err := s.SQLExecutor.QueryWith(builder)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*AppCollaborator
	for rows.Next() {
		r := &AppCollaborator{}
		err = rows.Scan(
			&r.AppID,
			&r.UserID,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, nil
}

func (s *GlobalDBStore) GetAppIDs() (appIDs []string, err error) {
	builder := s.SQLBuilder.
		Select(
			"app_id",
		).
		From(s.SQLBuilder.TableName("_portal_config_source")).
		OrderBy("created_at ASC")

	rows, e := s.SQLExecutor.QueryWith(builder)
	if e != nil {
		err = e
		return
	}
	defer rows.Close()
	for rows.Next() {
		var appID string
		err = rows.Scan(
			&appID,
		)
		if err != nil {
			return
		}
		appIDs = append(appIDs, appID)
	}
	return
}

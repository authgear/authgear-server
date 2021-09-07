package analytic

import (
	"time"

	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
)

type AppDBStore struct {
	SQLBuilder  *appdb.SQLBuilder
	SQLExecutor *appdb.SQLExecutor
}

func (s *AppDBStore) GetNewUserIDs(appID string, rangeFrom *time.Time, rangeTo *time.Time) ([]string, error) {
	builder := s.SQLBuilder.WithAppID(appID).
		Select(
			"id",
		).
		From(s.SQLBuilder.TableName("_auth_user")).
		Where("created_at >= ?", rangeFrom).
		Where("created_at < ?", rangeTo)
	rows, err := s.SQLExecutor.QueryWith(builder)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []string
	for rows.Next() {
		var userID string
		err = rows.Scan(
			&userID,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, userID)
	}
	return result, nil
}

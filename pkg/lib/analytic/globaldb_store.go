package analytic

import (
	"time"

	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
)

type GlobalDBStore struct {
	SQLBuilder  *globaldb.SQLBuilder
	SQLExecutor *globaldb.SQLExecutor
}

func (s *GlobalDBStore) GetNewAppOwners(rangeFrom *time.Time, rangeTo *time.Time) ([]*AppCollaborator, error) {
	builder := s.SQLBuilder.Global().
		Select(
			"app_id",
			"user_id",
		).
		From(s.SQLBuilder.TableName("_portal_app_collaborator")).
		Where("created_at >= ?", rangeFrom).
		Where("created_at < ?", rangeTo).
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

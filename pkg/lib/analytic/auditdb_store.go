package analytic

import (
	"time"

	"github.com/authgear/authgear-server/pkg/lib/infra/db/auditdb"
)

type AuditDBStore struct {
	SQLBuilder  *auditdb.SQLBuilder
	SQLExecutor *auditdb.ReadSQLExecutor
}

func (s *AuditDBStore) GetCountByActivityType(appID string, activityType string, rangeFrom *time.Time, rangeTo *time.Time) (int, error) {
	builder := s.SQLBuilder.WithAppID(appID).
		Select("count(*)").
		From(s.SQLBuilder.TableName("_audit_log")).
		Where("activity_type = ?", activityType).
		Where("created_at >= ?", rangeFrom).
		Where("created_at < ?", rangeTo)
	row, err := s.SQLExecutor.QueryRowWith(builder)
	if err != nil {
		return 0, err
	}
	var count int
	err = row.Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

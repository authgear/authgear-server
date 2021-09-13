package analytic

import (
	"time"

	"github.com/authgear/authgear-server/pkg/lib/infra/db/auditdb"
)

type AuditDBStore struct {
	SQLBuilder  *auditdb.SQLBuilder
	SQLExecutor *auditdb.WriteSQLExecutor
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

func (s *AuditDBStore) CreateCounts(counts []*Count) error {
	builder := s.SQLBuilder.Global().
		Insert(s.SQLBuilder.TableName("_audit_analytic_count")).
		Columns(
			"id",
			"app_id",
			"type",
			"count",
			"date",
		)

	for _, count := range counts {
		builder = builder.Values(
			count.ID,
			count.AppID,
			count.Type,
			count.Count,
			count.Date,
		)
	}
	_, err := s.SQLExecutor.ExecWith(builder)
	if err != nil {
		return err
	}

	return nil
}

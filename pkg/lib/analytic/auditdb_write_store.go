package analytic

import (
	"github.com/authgear/authgear-server/pkg/lib/infra/db/auditdb"
)

type AuditDBWriteStore struct {
	SQLBuilder  *auditdb.SQLBuilder
	SQLExecutor *auditdb.WriteSQLExecutor
}

// UpsertCounts upsert counts in batches
func (s *AuditDBWriteStore) UpsertCounts(counts []*Count) error {
	batchSize := 100
	for i := 0; i < len(counts); i += batchSize {
		j := i + batchSize
		if j > len(counts) {
			j = len(counts)
		}
		batch := counts[i:j]

		err := s.upsertCounts(batch)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *AuditDBWriteStore) upsertCounts(counts []*Count) error {
	builder := s.SQLBuilder.WithoutAppID().
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

	builder = builder.Suffix("ON CONFLICT (app_id, type, date) DO UPDATE SET count = excluded.count")
	_, err := s.SQLExecutor.ExecWith(builder)
	if err != nil {
		return err
	}

	return nil
}

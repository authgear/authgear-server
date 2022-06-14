package usage

import (
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
)

type GlobalDBStore struct {
	SQLBuilder  *globaldb.SQLBuilder
	SQLExecutor *globaldb.SQLExecutor
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

// UpsertUsageRecords upsert usage record in batches
func (s *GlobalDBStore) UpsertUsageRecords(usageRecords []*UsageRecord) error {
	batchSize := 100
	for i := 0; i < len(usageRecords); i += batchSize {
		j := i + batchSize
		if j > len(usageRecords) {
			j = len(usageRecords)
		}
		batch := usageRecords[i:j]

		err := s.upsertUsageRecords(batch)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *GlobalDBStore) upsertUsageRecords(usageRecords []*UsageRecord) error {
	builder := s.SQLBuilder.
		Insert(s.SQLBuilder.TableName("_portal_usage_record")).
		Columns(
			"id",
			"app_id",
			"name",
			"period",
			"start_time",
			"end_time",
			"count",
		)

	for _, u := range usageRecords {
		builder = builder.Values(
			u.ID,
			u.AppID,
			u.Name,
			u.Period,
			u.StartTime,
			u.EndTime,
			u.Count,
		)
	}

	builder = builder.Suffix("ON CONFLICT (app_id, name, period, start_time) DO UPDATE SET count = excluded.count RETURNING id")
	// TODO(usage): update id of usage record objects when conflict
	_, err := s.SQLExecutor.ExecWith(builder)
	if err != nil {
		return err
	}

	return nil
}

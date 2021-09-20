package analytic

import (
	"encoding/json"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/audit"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/auditdb"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type AuditDBReadStore struct {
	SQLBuilder  *auditdb.SQLBuilder
	SQLExecutor *auditdb.ReadSQLExecutor
}

func (s *AuditDBReadStore) GetCountByActivityType(appID string, activityType string, rangeFrom *time.Time, rangeTo *time.Time) (int, error) {
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

// QueryPage is copied from pkg/lib/audit/read_store.go
// The ReadStore cannot be used here as it requires appID during initialization through injection
func (s *AuditDBReadStore) QueryPage(appID string, opts audit.QueryPageOptions, pageArgs graphqlutil.PageArgs) ([]*audit.Log, uint64, error) {
	query := s.selectLogQuery(appID)

	query = opts.Apply(query)

	query = query.OrderBy("created_at ASC")

	query, offset, err := db.ApplyPageArgs(query, pageArgs)
	if err != nil {
		return nil, 0, err
	}

	rows, err := s.SQLExecutor.QueryWith(query)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []*audit.Log
	for rows.Next() {
		l, err := s.scanLog(rows)
		if err != nil {
			return nil, 0, err
		}
		logs = append(logs, l)
	}

	return logs, offset, nil
}

func (s *AuditDBReadStore) selectLogQuery(appID string) db.SelectBuilder {
	return s.SQLBuilder.WithAppID(appID).
		Select(
			"id",
			"created_at",
			"user_id",
			"activity_type",
			"ip_address",
			"user_agent",
			"client_id",
			"data",
		).
		From(s.SQLBuilder.TableName("_audit_log"))
}

func (s *AuditDBReadStore) scanLog(scn db.Scanner) (*audit.Log, error) {
	l := &audit.Log{}

	var data []byte

	err := scn.Scan(
		&l.ID,
		&l.CreatedAt,
		&l.UserID,
		&l.ActivityType,
		&l.IPAddress,
		&l.UserAgent,
		&l.ClientID,
		&data,
	)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &l.Data)
	if err != nil {
		return nil, err
	}

	return l, nil
}

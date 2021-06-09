package audit

import (
	"encoding/json"
	"time"

	"github.com/lib/pq"

	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/auditdb"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type QueryPageOptions struct {
	RangeFrom     *time.Time
	RangeTo       *time.Time
	ActivityTypes []string
}

func (o QueryPageOptions) Apply(q db.SelectBuilder) db.SelectBuilder {
	if o.RangeFrom != nil {
		q = q.Where("created_at >= ?", o.RangeFrom)
	}

	if o.RangeTo != nil {
		q = q.Where("created_at < ?", o.RangeTo)
	}

	if len(o.ActivityTypes) > 0 {
		q = q.Where("activity_type = ANY (?)", pq.Array(o.ActivityTypes))
	}

	return q
}

type Store struct {
	SQLBuilder  *auditdb.SQLBuilder
	SQLExecutor *auditdb.SQLExecutor
}

func (s *Store) PersistLog(logEntry *Log) (err error) {
	data, err := json.Marshal(logEntry.Data)
	if err != nil {
		return
	}

	builder := s.SQLBuilder.Tenant().
		Insert(s.SQLBuilder.TableName("_audit_log")).
		Columns(
			"id",
			"created_at",
			"user_id",
			"activity_type",
			"ip_address",
			"user_agent",
			"client_id",
			"data",
		).
		Values(
			logEntry.ID,
			logEntry.CreatedAt,
			logEntry.UserID,
			logEntry.ActivityType,
			logEntry.IPAddress,
			logEntry.UserAgent,
			logEntry.ClientID,
			data,
		)

	_, err = s.SQLExecutor.ExecWith(builder)
	if err != nil {
		return
	}

	return nil
}

func (s *Store) Count(opts QueryPageOptions) (uint64, error) {
	query := s.SQLBuilder.Tenant().
		Select("count(*)").
		From(s.SQLBuilder.TableName("_audit_log"))

	query = opts.Apply(query)

	scanner, err := s.SQLExecutor.QueryRowWith(query)
	if err != nil {
		return 0, err
	}

	var count uint64
	err = scanner.Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Store) QueryPage(opts QueryPageOptions, pageArgs graphqlutil.PageArgs) ([]*Log, uint64, error) {
	query := s.selectQuery()

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

	var logs []*Log
	for rows.Next() {
		l, err := s.scan(rows)
		if err != nil {
			return nil, 0, err
		}
		logs = append(logs, l)
	}

	return logs, offset, nil
}

func (s *Store) GetByIDs(ids []string) ([]*Log, error) {
	query := s.selectQuery().Where("id = ANY (?)", pq.Array(ids))

	rows, err := s.SQLExecutor.QueryWith(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*Log
	for rows.Next() {
		l, err := s.scan(rows)
		if err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}

	return logs, nil
}

func (s *Store) selectQuery() db.SelectBuilder {
	return s.SQLBuilder.Tenant().
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

func (s *Store) scan(scn db.Scanner) (*Log, error) {
	l := &Log{}

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

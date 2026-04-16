package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lib/pq"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/auditdb"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
	"github.com/authgear/authgear-server/pkg/util/slice"
)

type QueryPageOptions struct {
	RangeFrom      *time.Time
	RangeTo        *time.Time
	ActivityTypes  []string
	UserIDs        []string
	EmailAddresses []string
	PhoneNumbers   []string
	SortDirection  model.SortDirection
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

	q = o.applyQueryStringFilter(q)
	return q
}

func (o QueryPageOptions) applyQueryStringFilter(q db.SelectBuilder) db.SelectBuilder {
	hasUserIDs := len(o.UserIDs) > 0
	hasEmail := len(o.EmailAddresses) > 0
	hasPhone := len(o.PhoneNumbers) > 0

	mergedEmailsAndPhones := append([]string(nil), o.EmailAddresses...)
	mergedEmailsAndPhones = append(mergedEmailsAndPhones, o.PhoneNumbers...)
	mergedEmailsAndPhones = slice.Deduplicate(mergedEmailsAndPhones)

	switch {
	case hasUserIDs && hasEmail && hasPhone:
		q = q.Where("(user_id = ANY (?) OR data#>>'{payload,recipient}' = ANY (?))", pq.Array(o.UserIDs), pq.Array(mergedEmailsAndPhones))
	case hasUserIDs && hasEmail && !hasPhone:
		q = q.Where("(user_id = ANY (?) OR data#>>'{payload,recipient}' = ANY (?))", pq.Array(o.UserIDs), pq.Array(o.EmailAddresses))
	case hasUserIDs && !hasEmail && hasPhone:
		q = q.Where("(user_id = ANY (?) OR data#>>'{payload,recipient}' = ANY (?))", pq.Array(o.UserIDs), pq.Array(o.PhoneNumbers))
	case hasUserIDs && !hasEmail && !hasPhone:
		q = q.Where("user_id = ANY (?)", pq.Array(o.UserIDs))
	case !hasUserIDs && hasEmail && hasPhone:
		q = q.Where("data#>>'{payload,recipient}' = ANY (?)", pq.Array(mergedEmailsAndPhones))
	case !hasUserIDs && hasEmail && !hasPhone:
		q = q.Where("data#>>'{payload,recipient}' = ANY (?)", pq.Array(o.EmailAddresses))
	case !hasUserIDs && !hasEmail && hasPhone:
		q = q.Where("data#>>'{payload,recipient}' = ANY (?)", pq.Array(o.PhoneNumbers))
	case !hasUserIDs && !hasEmail && !hasPhone:
		fallthrough
	default:
		// do nothing
	}

	return q
}

type ReadStore struct {
	SQLBuilder  *auditdb.SQLBuilderApp
	SQLExecutor *auditdb.ReadSQLExecutor
}

func (s *ReadStore) Count(ctx context.Context, opts QueryPageOptions) (uint64, error) {
	query := s.SQLBuilder.
		Select("count(*)").
		From(s.SQLBuilder.TableName("_audit_log"))

	query = opts.Apply(query)

	scanner, err := s.SQLExecutor.QueryRowWith(ctx, query)
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

func (s *ReadStore) QueryPage(ctx context.Context, opts QueryPageOptions, pageArgs graphqlutil.PageArgs) ([]*Log, uint64, error) {
	query := s.selectQuery()

	query = opts.Apply(query)

	sortDirection := opts.SortDirection
	if sortDirection == model.SortDirectionDefault {
		sortDirection = model.SortDirectionDesc
	}

	query = query.OrderBy(fmt.Sprintf("created_at %s", sortDirection))

	query, offset, err := db.ApplyPageArgs(query, pageArgs)
	if err != nil {
		return nil, 0, err
	}

	rows, err := s.SQLExecutor.QueryWith(ctx, query)
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

func (s *ReadStore) GetByIDs(ctx context.Context, ids []string) ([]*Log, error) {
	query := s.selectQuery().Where("id = ANY (?)", pq.Array(ids))

	rows, err := s.SQLExecutor.QueryWith(ctx, query)
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

func (s *ReadStore) countFraudProtectionDecisionRecordsQuery(
	opts FraudProtectionDecisionRecordQueryOptions,
) db.SelectBuilder {
	query := s.SQLBuilder.
		Select("count(*)").
		From(s.SQLBuilder.TableName("_audit_log"))
	return opts.Apply(query)
}

func (s *ReadStore) queryFraudProtectionDecisionRecordsBase(
	opts FraudProtectionDecisionRecordQueryOptions,
) db.SelectBuilder {
	query := s.SQLBuilder.
		Select("id", "created_at", "data").
		From(s.SQLBuilder.TableName("_audit_log"))
	return opts.Apply(query)
}

func (s *ReadStore) CountFraudProtectionDecisionRecords(
	ctx context.Context,
	opts FraudProtectionDecisionRecordQueryOptions,
) (uint64, error) {
	query := s.countFraudProtectionDecisionRecordsQuery(opts)

	scanner, err := s.SQLExecutor.QueryRowWith(ctx, query)
	if err != nil {
		return 0, err
	}

	var count uint64
	if err := scanner.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (s *ReadStore) QueryFraudProtectionDecisionRecordsPage(
	ctx context.Context,
	opts FraudProtectionDecisionRecordQueryOptions,
	pageArgs graphqlutil.PageArgs,
) ([]*FraudProtectionDecisionRecord, uint64, error) {
	query := s.queryFraudProtectionDecisionRecordsBase(opts)

	sortDirection := opts.SortDirection
	if sortDirection == model.SortDirectionDefault {
		sortDirection = model.SortDirectionDesc
	}
	query = query.OrderBy(fmt.Sprintf("created_at %s", sortDirection))

	query, offset, err := db.ApplyPageArgs(query, pageArgs)
	if err != nil {
		return nil, 0, err
	}

	rows, err := s.SQLExecutor.QueryWith(ctx, query)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var records []*FraudProtectionDecisionRecord
	for rows.Next() {
		record, err := s.scanFraudProtectionDecisionRecord(rows)
		if err != nil {
			return nil, 0, err
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return records, offset, nil
}

func (s *ReadStore) GetFraudProtectionDecisionRecordByID(
	ctx context.Context,
	id string,
) (*FraudProtectionDecisionRecord, error) {
	query := s.queryFraudProtectionDecisionRecordsBase(FraudProtectionDecisionRecordQueryOptions{}).
		Where("id = ?", id).
		Limit(1)

	row, err := s.SQLExecutor.QueryRowWith(ctx, query)
	if err != nil {
		return nil, err
	}

	record, err := s.scanFraudProtectionDecisionRecord(row)
	if err != nil {
		return nil, err
	}
	return record, nil
}

func (s *ReadStore) selectQuery() db.SelectBuilder {
	return s.SQLBuilder.
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

func (s *ReadStore) scan(scn db.Scanner) (*Log, error) {
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

func (s *ReadStore) scanFraudProtectionDecisionRecord(
	scn db.Scanner,
) (*FraudProtectionDecisionRecord, error) {
	record := &FraudProtectionDecisionRecord{}
	var raw []byte
	if err := scn.Scan(&record.ID, &record.CreatedAt, &raw); err != nil {
		return nil, err
	}

	var payload struct {
		Payload struct {
			Record model.FraudProtectionDecisionRecord `json:"record"`
		} `json:"payload"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, err
	}
	record.Record = payload.Payload.Record
	return record, nil
}

package analytic

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/auditdb"
)

type AuditDBReadStore struct {
	SQLBuilder  *auditdb.SQLBuilder
	SQLExecutor *auditdb.ReadSQLExecutor
}

func (s *AuditDBReadStore) GetAnalyticCountByType(
	ctx context.Context,
	appID string,
	typ string,
	date *time.Time,
) (*Count, error) {
	builder := s.selectAnalyticCountQuery(appID).
		Where("type = ?", typ).
		Where("date = ?", date)
	row, err := s.SQLExecutor.QueryRowWith(ctx, builder)
	if err != nil {
		return nil, err
	}

	count, err := s.scanAnalyticCount(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrAnalyticCountNotFound
	} else if err != nil {
		return nil, err
	}
	return count, nil
}

// GetAnalyticCountsByType get counts by type and date range
// the provided rangeFrom and rangeTo are inclusive
func (s *AuditDBReadStore) GetAnalyticCountsByType(
	ctx context.Context,
	appID string,
	typ string,
	rangeFrom *time.Time,
	rangeTo *time.Time,
) ([]*Count, error) {
	builder := s.selectAnalyticCountQuery(appID).
		Where("type = ?", typ).
		Where("date >= ?", rangeFrom).
		Where("date <= ?", rangeTo)

	rows, err := s.SQLExecutor.QueryWith(ctx, builder)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var counts []*Count
	for rows.Next() {
		count, err := s.scanAnalyticCount(rows)
		if err != nil {
			return nil, err
		}
		counts = append(counts, count)
	}

	return counts, nil
}

func (s *AuditDBReadStore) GetSumOfAnalyticCountsByType(
	ctx context.Context,
	appID string,
	typ string,
	rangeFrom *time.Time,
	rangeTo *time.Time,
) (int, error) {
	builder := s.SQLBuilder.WithAppID(appID).
		Select(
			"sum(count)",
		).
		From(s.SQLBuilder.TableName("_audit_analytic_count")).
		Where("type = ?", typ).
		Where("date >= ?", rangeFrom).
		Where("date <= ?", rangeTo)

	row, err := s.SQLExecutor.QueryRowWith(ctx, builder)
	if err != nil {
		return 0, err
	}

	var sum sql.NullInt64
	err = row.Scan(&sum)
	if err != nil {
		return 0, err
	}

	if sum.Valid {
		return int(sum.Int64), nil
	}

	return 0, nil
}

func (s *AuditDBReadStore) selectAnalyticCountQuery(appID string) db.SelectBuilder {
	return s.SQLBuilder.WithAppID(appID).
		Select(
			"id",
			"app_id",
			"count",
			"date",
			"type",
		).
		From(s.SQLBuilder.TableName("_audit_analytic_count"))
}

func (s *AuditDBReadStore) scanAnalyticCount(scn db.Scanner) (*Count, error) {
	c := &Count{}
	err := scn.Scan(
		&c.ID,
		&c.AppID,
		&c.Count,
		&c.Date,
		&c.Type,
	)
	if err != nil {
		return nil, err
	}
	return c, nil
}

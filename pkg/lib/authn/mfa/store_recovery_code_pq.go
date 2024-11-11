package mfa

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"

	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
)

type StoreRecoveryCodePQ struct {
	SQLBuilder  *appdb.SQLBuilderApp
	SQLExecutor *appdb.SQLExecutor
}

func (s *StoreRecoveryCodePQ) List(ctx context.Context, userID string) ([]*RecoveryCode, error) {
	builder := s.SQLBuilder.
		Select(
			"rc.id",
			"rc.user_id",
			"rc.code",
			"rc.created_at",
			"rc.updated_at",
			"rc.consumed",
		).
		From(s.SQLBuilder.TableName("_auth_recovery_code"), "rc").
		Where("rc.user_id = ?", userID)

	rows, err := s.SQLExecutor.QueryWith(ctx, builder)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var codes []*RecoveryCode
	for rows.Next() {
		rc := &RecoveryCode{}
		err = rows.Scan(
			&rc.ID,
			&rc.UserID,
			&rc.Code,
			&rc.CreatedAt,
			&rc.UpdatedAt,
			&rc.Consumed,
		)
		if err != nil {
			return nil, err
		}
		codes = append(codes, rc)
	}

	return codes, nil
}

func (s *StoreRecoveryCodePQ) Get(ctx context.Context, userID string, code string) (*RecoveryCode, error) {
	builder := s.SQLBuilder.
		Select(
			"rc.id",
			"rc.user_id",
			"rc.code",
			"rc.created_at",
			"rc.updated_at",
			"rc.consumed",
		).
		From(s.SQLBuilder.TableName("_auth_recovery_code"), "rc").
		Where("rc.user_id = ? AND rc.code = ?", userID, code)

	row, err := s.SQLExecutor.QueryRowWith(ctx, builder)
	if err != nil {
		return nil, err
	}

	rc := &RecoveryCode{}
	err = row.Scan(
		&rc.ID,
		&rc.UserID,
		&rc.Code,
		&rc.CreatedAt,
		&rc.UpdatedAt,
		&rc.Consumed,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrRecoveryCodeNotFound
	} else if err != nil {
		return nil, err
	}

	return rc, nil
}

func (s *StoreRecoveryCodePQ) DeleteAll(ctx context.Context, userID string) error {
	ids, err := func() ([]string, error) {
		builder := s.SQLBuilder.
			Select("rc.id").
			From(s.SQLBuilder.TableName("_auth_recovery_code"), "rc").
			Where("rc.user_id = ?", userID)

		rows, err := s.SQLExecutor.QueryWith(ctx, builder)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var ids []string
		for rows.Next() {
			var id string
			err = rows.Scan(&id)
			if err != nil {
				return nil, err
			}
			ids = append(ids, id)
		}
		return ids, nil
	}()
	if err != nil {
		return err
	}

	if len(ids) == 0 {
		return nil
	}

	q := s.SQLBuilder.
		Delete(s.SQLBuilder.TableName("_auth_recovery_code")).
		Where("id = ANY (?)", pq.Array(ids))
	_, err = s.SQLExecutor.ExecWith(ctx, q)
	if err != nil {
		return err
	}

	return nil
}

func (s *StoreRecoveryCodePQ) CreateAll(ctx context.Context, codes []*RecoveryCode) error {
	q := s.SQLBuilder.
		Insert(s.SQLBuilder.TableName("_auth_recovery_code")).
		Columns(
			"id",
			"user_id",
			"code",
			"created_at",
			"updated_at",
			"consumed",
		)

	for _, a := range codes {
		q = q.Values(
			a.ID,
			a.UserID,
			a.Code,
			a.CreatedAt,
			a.UpdatedAt,
			a.Consumed,
		)
	}

	_, err := s.SQLExecutor.ExecWith(ctx, q)
	if err != nil {
		return err
	}

	return nil
}

func (s *StoreRecoveryCodePQ) UpdateConsumed(ctx context.Context, code *RecoveryCode) error {
	q := s.SQLBuilder.
		Update(s.SQLBuilder.TableName("_auth_recovery_code")).
		Where("id = ?", code.ID).
		Set("consumed", code.Consumed).
		Set("updated_at", code.UpdatedAt)
	_, err := s.SQLExecutor.ExecWith(ctx, q)
	if err != nil {
		return err
	}

	return nil
}

package mfa

import (
	"database/sql"
	"errors"

	"github.com/lib/pq"

	"github.com/authgear/authgear-server/pkg/db"
)

type StoreRecoveryCodePQ struct {
	SQLBuilder  db.SQLBuilder
	SQLExecutor db.SQLExecutor
}

func (s *StoreRecoveryCodePQ) List(userID string) ([]*RecoveryCode, error) {
	builder := s.SQLBuilder.Tenant().
		Select(
			"rc.id",
			"rc.user_id",
			"rc.code",
			"rc.created_at",
			"rc.consumed",
		).
		From(s.SQLBuilder.FullTableName("recovery_code"), "rc").
		Where("rc.user_id = ?", userID)

	rows, err := s.SQLExecutor.QueryWith(builder)
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
			&rc.Consumed,
		)
		if err != nil {
			return nil, err
		}
		codes = append(codes, rc)
	}

	return codes, nil
}

func (s *StoreRecoveryCodePQ) Get(userID string, code string) (*RecoveryCode, error) {
	builder := s.SQLBuilder.Tenant().
		Select(
			"rc.id",
			"rc.user_id",
			"rc.code",
			"rc.created_at",
			"rc.consumed",
		).
		From(s.SQLBuilder.FullTableName("recovery_code"), "rc").
		Where("rc.user_id = ? AND rc.code = ?", userID, code)

	row, err := s.SQLExecutor.QueryRowWith(builder)
	if err != nil {
		return nil, err
	}

	rc := &RecoveryCode{}
	err = row.Scan(
		&rc.ID,
		&rc.UserID,
		&rc.Code,
		&rc.CreatedAt,
		&rc.Consumed,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrRecoveryCodeNotFound
	} else if err != nil {
		return nil, err
	}

	return rc, nil
}

func (s *StoreRecoveryCodePQ) DeleteAll(userID string) error {
	ids, err := func() ([]string, error) {
		builder := s.SQLBuilder.Tenant().
			Select("rc.id").
			From(s.SQLBuilder.FullTableName("recovery_code"), "rc").
			Where("rc.user_id = ?", userID)

		rows, err := s.SQLExecutor.QueryWith(builder)
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

	q := s.SQLBuilder.Tenant().
		Delete(s.SQLBuilder.FullTableName("recovery_code")).
		Where("id = ANY (?)", pq.Array(ids))
	_, err = s.SQLExecutor.ExecWith(q)
	if err != nil {
		return err
	}

	return nil
}

func (s *StoreRecoveryCodePQ) CreateAll(codes []*RecoveryCode) error {
	q := s.SQLBuilder.Tenant().
		Insert(s.SQLBuilder.FullTableName("recovery_code")).
		Columns(
			"id",
			"user_id",
			"code",
			"created_at",
			"consumed",
		)

	for _, a := range codes {
		q = q.Values(
			a.ID,
			a.UserID,
			a.Code,
			a.CreatedAt,
			a.Consumed,
		)
	}

	_, err := s.SQLExecutor.ExecWith(q)
	if err != nil {
		return err
	}

	return nil
}

func (s *StoreRecoveryCodePQ) MarkConsumed(code *RecoveryCode) error {
	q := s.SQLBuilder.Tenant().
		Update(s.SQLBuilder.FullTableName("recovery_code")).
		Where("id = ?", code.ID).
		Set("consumed", true)
	_, err := s.SQLExecutor.ExecWith(q)
	if err != nil {
		return err
	}

	code.Consumed = true
	return nil
}

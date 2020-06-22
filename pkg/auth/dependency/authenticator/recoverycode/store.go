package recoverycode

import (
	"database/sql"
	"errors"

	"github.com/lib/pq"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authenticator"
	"github.com/skygeario/skygear-server/pkg/core/authn"
	"github.com/skygeario/skygear-server/pkg/db"
)

type Store struct {
	SQLBuilder  db.SQLBuilder
	SQLExecutor db.SQLExecutor
}

func (s *Store) Get(userID string, id string) (*Authenticator, error) {
	builder := s.SQLBuilder.Tenant().
		Select(
			"a.id",
			"a.user_id",
			"arc.code",
			"arc.created_at",
			"arc.consumed",
		).
		From(s.SQLBuilder.FullTableName("authenticator"), "a").
		Join(
			s.SQLBuilder.FullTableName("authenticator_recovery_code"),
			"arc",
			"a.id = arc.id",
		).
		Where("a.id = ? AND a.user_id = ?", id, userID)

	row, err := s.SQLExecutor.QueryRowWith(builder)
	if err != nil {
		return nil, err
	}

	a := &Authenticator{}
	err = row.Scan(
		&a.ID,
		&a.UserID,
		&a.Code,
		&a.CreatedAt,
		&a.Consumed,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, authenticator.ErrAuthenticatorNotFound
	} else if err != nil {
		return nil, err
	}

	return a, nil
}

func (s *Store) List(userID string) ([]*Authenticator, error) {
	builder := s.SQLBuilder.Tenant().
		Select(
			"a.id",
			"a.user_id",
			"arc.code",
			"arc.created_at",
			"arc.consumed",
		).
		From(s.SQLBuilder.FullTableName("authenticator"), "a").
		Join(
			s.SQLBuilder.FullTableName("authenticator_recovery_code"),
			"arc",
			"a.id = arc.id",
		).
		Where("a.user_id = ?", userID)

	rows, err := s.SQLExecutor.QueryWith(builder)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var authenticators []*Authenticator
	for rows.Next() {
		a := &Authenticator{}
		err = rows.Scan(
			&a.ID,
			&a.UserID,
			&a.Code,
			&a.CreatedAt,
			&a.Consumed,
		)
		if err != nil {
			return nil, err
		}
		authenticators = append(authenticators, a)
	}

	return authenticators, nil
}

func (s *Store) DeleteAll(userID string) error {
	ids, err := func() ([]string, error) {
		builder := s.SQLBuilder.Tenant().
			Select("a.id").
			From(s.SQLBuilder.FullTableName("authenticator"), "a").
			Where("type = ? AND a.user_id = ?", authn.AuthenticatorTypeRecoveryCode, userID)

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
		Delete(s.SQLBuilder.FullTableName("authenticator_recovery_code")).
		Where("id = ANY (?)", pq.Array(ids))
	_, err = s.SQLExecutor.ExecWith(q)
	if err != nil {
		return err
	}

	q = s.SQLBuilder.Tenant().
		Delete(s.SQLBuilder.FullTableName("authenticator")).
		Where("id = ANY (?)", pq.Array(ids))
	_, err = s.SQLExecutor.ExecWith(q)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) CreateAll(authenticators []*Authenticator) error {
	q1 := s.SQLBuilder.Tenant().
		Insert(s.SQLBuilder.FullTableName("authenticator")).
		Columns(
			"id",
			"type",
			"user_id",
		)

	q2 := s.SQLBuilder.Tenant().
		Insert(s.SQLBuilder.FullTableName("authenticator_recovery_code")).
		Columns(
			"id",
			"code",
			"created_at",
			"consumed",
		)

	for _, a := range authenticators {
		q1 = q1.Values(
			a.ID,
			authn.AuthenticatorTypeRecoveryCode,
			a.UserID,
		)
		q2 = q2.Values(
			a.ID,
			a.Code,
			a.CreatedAt,
			a.Consumed,
		)
	}

	_, err := s.SQLExecutor.ExecWith(q1)
	if err != nil {
		return err
	}
	_, err = s.SQLExecutor.ExecWith(q2)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) MarkConsumed(authenticator *Authenticator) error {
	q := s.SQLBuilder.Tenant().
		Update(s.SQLBuilder.FullTableName("authenticator_recovery_code")).
		Where("id = ?", authenticator.ID).
		Set("consumed", true)
	_, err := s.SQLExecutor.ExecWith(q)
	if err != nil {
		return err
	}

	authenticator.Consumed = true
	return nil
}

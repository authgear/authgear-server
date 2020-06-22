package password

import (
	"database/sql"
	"errors"

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
			"ap.password_hash",
		).
		From(s.SQLBuilder.FullTableName("authenticator"), "a").
		Join(
			s.SQLBuilder.FullTableName("authenticator_password"),
			"ap",
			"a.id = ap.id",
		).
		Where("a.user_id = ? AND a.id = ?", userID, id)

	row, err := s.SQLExecutor.QueryRowWith(builder)
	if err != nil {
		return nil, err
	}

	a := &Authenticator{}
	err = row.Scan(
		&a.ID,
		&a.UserID,
		&a.PasswordHash,
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
			"ap.password_hash",
		).
		From(s.SQLBuilder.FullTableName("authenticator"), "a").
		Join(
			s.SQLBuilder.FullTableName("authenticator_password"),
			"ap",
			"a.id = ap.id",
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
			&a.PasswordHash,
		)
		if err != nil {
			return nil, err
		}
		authenticators = append(authenticators, a)
	}

	return authenticators, nil
}

func (s *Store) Delete(id string) error {
	q := s.SQLBuilder.Tenant().
		Delete(s.SQLBuilder.FullTableName("authenticator_password")).
		Where("id = ?", id)
	_, err := s.SQLExecutor.ExecWith(q)
	if err != nil {
		return err
	}

	q = s.SQLBuilder.Tenant().
		Delete(s.SQLBuilder.FullTableName("authenticator")).
		Where("id = ?", id)
	_, err = s.SQLExecutor.ExecWith(q)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) Create(a *Authenticator) error {
	q := s.SQLBuilder.Tenant().
		Insert(s.SQLBuilder.FullTableName("authenticator")).
		Columns(
			"id",
			"type",
			"user_id",
		).
		Values(
			a.ID,
			authn.AuthenticatorTypePassword,
			a.UserID,
		)
	_, err := s.SQLExecutor.ExecWith(q)
	if err != nil {
		return err
	}

	q = s.SQLBuilder.Tenant().
		Insert(s.SQLBuilder.FullTableName("authenticator_password")).
		Columns(
			"id",
			"password_hash",
		).
		Values(
			a.ID,
			a.PasswordHash,
		)
	_, err = s.SQLExecutor.ExecWith(q)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) UpdatePasswordHash(a *Authenticator) error {
	q := s.SQLBuilder.Tenant().
		Update(s.SQLBuilder.FullTableName("authenticator_password")).
		Set("password_hash", a.PasswordHash).
		Where("id = ?", a.ID)
	_, err := s.SQLExecutor.ExecWith(q)
	if err != nil {
		return err
	}

	return nil
}

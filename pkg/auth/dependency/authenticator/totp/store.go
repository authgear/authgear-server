package totp

import (
	"database/sql"
	"errors"

	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/core/authn"
	"github.com/authgear/authgear-server/pkg/db"
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
			"at.created_at",
			"at.secret",
			"at.display_name",
		).
		From(s.SQLBuilder.FullTableName("authenticator"), "a").
		Join(
			s.SQLBuilder.FullTableName("authenticator_totp"),
			"at",
			"a.id = at.id",
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
		&a.CreatedAt,
		&a.Secret,
		&a.DisplayName,
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
			"at.created_at",
			"at.secret",
			"at.display_name",
		).
		From(s.SQLBuilder.FullTableName("authenticator"), "a").
		Join(
			s.SQLBuilder.FullTableName("authenticator_totp"),
			"at",
			"a.id = at.id",
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
			&a.CreatedAt,
			&a.Secret,
			&a.DisplayName,
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
		Delete(s.SQLBuilder.FullTableName("authenticator_totp")).
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
			authn.AuthenticatorTypeTOTP,
			a.UserID,
		)
	_, err := s.SQLExecutor.ExecWith(q)
	if err != nil {
		return err
	}

	q = s.SQLBuilder.Tenant().
		Insert(s.SQLBuilder.FullTableName("authenticator_totp")).
		Columns(
			"id",
			"created_at",
			"secret",
			"display_name",
		).
		Values(
			a.ID,
			a.CreatedAt,
			a.Secret,
			a.DisplayName,
		)
	_, err = s.SQLExecutor.ExecWith(q)
	if err != nil {
		return err
	}

	return nil
}

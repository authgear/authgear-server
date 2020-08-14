package password

import (
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
)

type Store struct {
	SQLBuilder  db.SQLBuilder
	SQLExecutor db.SQLExecutor
}

func (s *Store) selectQuery() db.SelectBuilder {
	return s.SQLBuilder.Tenant().
		Select(
			"a.id",
			"a.user_id",
			"a.tag",
			"ap.password_hash",
		).
		From(s.SQLBuilder.FullTableName("authenticator"), "a").
		Join(s.SQLBuilder.FullTableName("authenticator_password"), "ap", "a.id = ap.id")
}

func (s *Store) scan(scn db.Scanner) (*Authenticator, error) {
	a := &Authenticator{}
	var tag []byte

	err := scn.Scan(
		&a.ID,
		&a.UserID,
		&tag,
		&a.PasswordHash,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, authenticator.ErrAuthenticatorNotFound
	} else if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(tag, &a.Tag); err != nil {
		return nil, err
	}

	return a, nil
}

func (s *Store) Get(userID string, id string) (*Authenticator, error) {
	q := s.selectQuery().Where("a.user_id = ? AND a.id = ?", userID, id)

	row, err := s.SQLExecutor.QueryRowWith(q)
	if err != nil {
		return nil, err
	}

	return s.scan(row)
}

func (s *Store) List(userID string) ([]*Authenticator, error) {
	q := s.selectQuery().Where("a.user_id = ?", userID)

	rows, err := s.SQLExecutor.QueryWith(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var authenticators []*Authenticator
	for rows.Next() {
		a, err := s.scan(rows)
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
	tag, err := json.Marshal(a.Tag)
	if err != nil {
		return err
	}

	q := s.SQLBuilder.Tenant().
		Insert(s.SQLBuilder.FullTableName("authenticator")).
		Columns(
			"id",
			"type",
			"user_id",
			"tag",
		).
		Values(
			a.ID,
			authn.AuthenticatorTypePassword,
			a.UserID,
			tag,
		)
	_, err = s.SQLExecutor.ExecWith(q)
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

package password

import (
	"database/sql"
	"errors"

	"github.com/lib/pq"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
)

type Store struct {
	SQLBuilder  *appdb.SQLBuilderApp
	SQLExecutor *appdb.SQLExecutor
}

func (s *Store) selectQuery() db.SelectBuilder {
	return s.SQLBuilder.
		Select(
			"a.id",
			"a.user_id",
			"a.created_at",
			"a.updated_at",
			"a.is_default",
			"a.kind",
			"ap.password_hash",
			"ap.expire_after",
		).
		From(s.SQLBuilder.TableName("_auth_authenticator"), "a").
		Join(s.SQLBuilder.TableName("_auth_authenticator_password"), "ap", "a.id = ap.id")
}

func (s *Store) scan(scn db.Scanner) (*authenticator.Password, error) {
	a := &authenticator.Password{}

	err := scn.Scan(
		&a.ID,
		&a.UserID,
		&a.CreatedAt,
		&a.UpdatedAt,
		&a.IsDefault,
		&a.Kind,
		&a.PasswordHash,
		&a.ExpireAfter,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, authenticator.ErrAuthenticatorNotFound
	} else if err != nil {
		return nil, err
	}

	return a, nil
}

func (s *Store) GetMany(ids []string) ([]*authenticator.Password, error) {
	builder := s.selectQuery().Where("a.id = ANY (?)", pq.Array(ids))

	rows, err := s.SQLExecutor.QueryWith(builder)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var as []*authenticator.Password
	for rows.Next() {
		a, err := s.scan(rows)
		if err != nil {
			return nil, err
		}
		as = append(as, a)
	}

	return as, nil
}

func (s *Store) Get(userID string, id string) (*authenticator.Password, error) {
	q := s.selectQuery().Where("a.user_id = ? AND a.id = ?", userID, id)

	row, err := s.SQLExecutor.QueryRowWith(q)
	if err != nil {
		return nil, err
	}

	return s.scan(row)
}

func (s *Store) List(userID string) ([]*authenticator.Password, error) {
	q := s.selectQuery().Where("a.user_id = ?", userID)

	rows, err := s.SQLExecutor.QueryWith(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var authenticators []*authenticator.Password
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
	q := s.SQLBuilder.
		Delete(s.SQLBuilder.TableName("_auth_authenticator_password")).
		Where("id = ?", id)
	_, err := s.SQLExecutor.ExecWith(q)
	if err != nil {
		return err
	}

	q = s.SQLBuilder.
		Delete(s.SQLBuilder.TableName("_auth_authenticator")).
		Where("id = ?", id)
	_, err = s.SQLExecutor.ExecWith(q)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) Create(a *authenticator.Password) (err error) {
	q := s.SQLBuilder.
		Insert(s.SQLBuilder.TableName("_auth_authenticator")).
		Columns(
			"id",
			"type",
			"user_id",
			"created_at",
			"updated_at",
			"is_default",
			"kind",
		).
		Values(
			a.ID,
			model.AuthenticatorTypePassword,
			a.UserID,
			a.CreatedAt,
			a.UpdatedAt,
			a.IsDefault,
			a.Kind,
		)
	_, err = s.SQLExecutor.ExecWith(q)
	if err != nil {
		return err
	}

	q = s.SQLBuilder.
		Insert(s.SQLBuilder.TableName("_auth_authenticator_password")).
		Columns(
			"id",
			"password_hash",
			"expire_after",
		).
		Values(
			a.ID,
			a.PasswordHash,
			a.ExpireAfter,
		)
	_, err = s.SQLExecutor.ExecWith(q)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) UpdatePasswordHash(a *authenticator.Password) error {
	q := s.SQLBuilder.
		Update(s.SQLBuilder.TableName("_auth_authenticator_password")).
		Set("password_hash", a.PasswordHash).
		Where("id = ?", a.ID)
	_, err := s.SQLExecutor.ExecWith(q)
	if err != nil {
		return err
	}

	q = s.SQLBuilder.
		Update(s.SQLBuilder.TableName("_auth_authenticator")).
		Set("updated_at", a.UpdatedAt).
		Where("id = ?", a.ID)

	_, err = s.SQLExecutor.ExecWith(q)
	if err != nil {
		return err
	}

	return nil
}

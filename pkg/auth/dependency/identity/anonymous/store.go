package anonymous

import (
	"database/sql"
	"errors"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity"
	"github.com/skygeario/skygear-server/pkg/core/authn"
	"github.com/skygeario/skygear-server/pkg/core/db"
)

type Store struct {
	SQLBuilder  db.SQLBuilder
	SQLExecutor db.SQLExecutor
}

func (s *Store) selectQuery() db.SelectBuilder {
	return s.SQLBuilder.Tenant().
		Select(
			"p.id",
			"p.user_id",
			"a.key_id",
			"a.key",
		).
		From(s.SQLBuilder.FullTableName("identity"), "p").
		Join(s.SQLBuilder.FullTableName("identity_anonymous"), "a", "p.id = a.id")
}

func (s *Store) scan(scn db.Scanner) (*Identity, error) {
	i := &Identity{}
	err := scn.Scan(
		&i.ID,
		&i.UserID,
		&i.KeyID,
		&i.Key,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, identity.ErrIdentityNotFound
	} else if err != nil {
		return nil, err
	}

	return i, nil
}

func (s *Store) List(userID string) ([]*Identity, error) {
	q := s.selectQuery().Where("p.user_id = ?", userID)

	rows, err := s.SQLExecutor.QueryWith(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var is []*Identity
	for rows.Next() {
		i, err := s.scan(rows)
		if err != nil {
			return nil, err
		}
		is = append(is, i)
	}

	return is, nil
}

func (s *Store) ListByClaim(name string, value string) ([]*Identity, error) {
	if name != "kid" {
		return nil, nil
	}

	q := s.selectQuery().Where("a.key_id = ?", value)

	rows, err := s.SQLExecutor.QueryWith(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var is []*Identity
	for rows.Next() {
		i, err := s.scan(rows)
		if err != nil {
			return nil, err
		}
		is = append(is, i)
	}

	return is, nil
}

func (s *Store) Get(userID, id string) (*Identity, error) {
	q := s.selectQuery().Where("p.user_id = ? AND p.id = ?", userID, id)
	rows, err := s.SQLExecutor.QueryRowWith(q)
	if err != nil {
		return nil, err
	}

	return s.scan(rows)
}

func (s *Store) GetByKeyID(keyID string) (*Identity, error) {
	q := s.selectQuery().Where("a.key_id = ?", keyID)
	rows, err := s.SQLExecutor.QueryRowWith(q)
	if err != nil {
		return nil, err
	}

	return s.scan(rows)
}

func (s *Store) Create(i *Identity) error {
	builder := s.SQLBuilder.Tenant().
		Insert(s.SQLBuilder.FullTableName("identity")).
		Columns(
			"id",
			"type",
			"user_id",
		).
		Values(
			i.ID,
			authn.IdentityTypeAnonymous,
			i.UserID,
		)

	_, err := s.SQLExecutor.ExecWith(builder)
	if err != nil {
		return err
	}

	q := s.SQLBuilder.Tenant().
		Insert(s.SQLBuilder.FullTableName("identity_anonymous")).
		Columns(
			"id",
			"key_id",
			"key",
		).
		Values(
			i.ID,
			i.KeyID,
			i.Key,
		)

	_, err = s.SQLExecutor.ExecWith(q)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) Delete(i *Identity) error {
	q := s.SQLBuilder.Tenant().
		Delete(s.SQLBuilder.FullTableName("identity_anonymous")).
		Where("id = ?", i.ID)

	_, err := s.SQLExecutor.ExecWith(q)
	if err != nil {
		return err
	}

	q = s.SQLBuilder.Tenant().
		Delete(s.SQLBuilder.FullTableName("identity")).
		Where("id = ?", i.ID)

	_, err = s.SQLExecutor.ExecWith(q)
	if err != nil {
		return err
	}

	return nil
}

package loginid

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/lib/pq"
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
			"l.login_id_key",
			"l.login_id",
			"l.original_login_id",
			"l.unique_key",
			"l.claims",
		).
		From(s.SQLBuilder.FullTableName("identity"), "p").
		Join(s.SQLBuilder.FullTableName("identity_login_id"), "l", "p.id = l.id")
}

func (s *Store) scan(scn db.Scanner) (*Identity, error) {
	i := &Identity{}
	var claims []byte

	err := scn.Scan(
		&i.ID,
		&i.UserID,
		&i.LoginIDKey,
		&i.LoginID,
		&i.OriginalLoginID,
		&i.UniqueKey,
		&claims,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, identity.ErrIdentityNotFound
	} else if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(claims, &i.Claims); err != nil {
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
	q := s.selectQuery().
		Where("(l.claims #>> ?) = ?", pq.Array([]string{name}), value)

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

func (s *Store) GetByLoginID(loginIDKey string, loginID string) (*Identity, error) {
	q := s.selectQuery().Where(`l.login_id = ? AND l.login_id_key = ?`, loginID, loginIDKey)
	rows, err := s.SQLExecutor.QueryRowWith(q)
	if err != nil {
		return nil, err
	}

	return s.scan(rows)
}

func (s *Store) GetByUniqueKey(uniqueKey string) (*Identity, error) {
	q := s.selectQuery().Where(`l.unique_key = ?`, uniqueKey)
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
			authn.IdentityTypeLoginID,
			i.UserID,
		)

	_, err := s.SQLExecutor.ExecWith(builder)
	if err != nil {
		return err
	}

	claims, err := json.Marshal(i.Claims)
	if err != nil {
		return err
	}

	q := s.SQLBuilder.Tenant().
		Insert(s.SQLBuilder.FullTableName("identity_login_id")).
		Columns(
			"id",
			"login_id_key",
			"login_id",
			"original_login_id",
			"unique_key",
			"claims",
		).
		Values(
			i.ID,
			i.LoginIDKey,
			i.LoginID,
			i.OriginalLoginID,
			i.UniqueKey,
			claims,
		)

	_, err = s.SQLExecutor.ExecWith(q)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) Update(i *Identity) error {
	claims, err := json.Marshal(i.Claims)
	if err != nil {
		return err
	}

	q := s.SQLBuilder.Tenant().
		Update(s.SQLBuilder.FullTableName("identity_login_id")).
		Set("login_id_key", i.LoginIDKey).
		Set("login_id", i.LoginID).
		Set("original_login_id", i.OriginalLoginID).
		Set("unique_key", i.UniqueKey).
		Set("claims", claims).
		Where("id = ?", i.ID)

	result, err := s.SQLExecutor.ExecWith(q)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return identity.ErrIdentityNotFound
	} else if rowsAffected > 1 {
		panic(fmt.Sprintf("identity_oauth: want 1 row updated, got %v", rowsAffected))
	}

	return nil
}

func (s *Store) Delete(i *Identity) error {
	q := s.SQLBuilder.Tenant().
		Delete(s.SQLBuilder.FullTableName("identity_login_id")).
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

package loginid

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/lib/pq"

	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
)

type Store struct {
	SQLBuilder  *appdb.SQLBuilder
	SQLExecutor *appdb.SQLExecutor
}

func (s *Store) selectQuery() db.SelectBuilder {
	return s.SQLBuilder.Tenant().
		Select(
			"p.id",
			"p.labels",
			"p.user_id",
			"p.created_at",
			"p.updated_at",
			"l.login_id_key",
			"l.login_id_type",
			"l.login_id",
			"l.original_login_id",
			"l.unique_key",
			"l.claims",
		).
		From(s.SQLBuilder.TableName("_auth_identity"), "p").
		Join(s.SQLBuilder.TableName("_auth_identity_login_id"), "l", "p.id = l.id")
}

func (s *Store) scan(scn db.Scanner) (*Identity, error) {
	i := &Identity{}
	var labels []byte
	var claims []byte

	err := scn.Scan(
		&i.ID,
		&labels,
		&i.UserID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.LoginIDKey,
		&i.LoginIDType,
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

	if err = json.Unmarshal(labels, &i.Labels); err != nil {
		return nil, err
	}

	if err = json.Unmarshal(claims, &i.Claims); err != nil {
		return nil, err
	}

	return i, nil
}

func (s *Store) GetMany(ids []string) ([]*Identity, error) {
	builder := s.selectQuery().Where("p.id = ANY (?)", pq.Array(ids))

	rows, err := s.SQLExecutor.QueryWith(builder)
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
	labels, err := json.Marshal(i.Labels)
	if err != nil {
		return err
	}

	builder := s.SQLBuilder.Tenant().
		Insert(s.SQLBuilder.TableName("_auth_identity")).
		Columns(
			"id",
			"labels",
			"type",
			"user_id",
			"created_at",
			"updated_at",
		).
		Values(
			i.ID,
			labels,
			authn.IdentityTypeLoginID,
			i.UserID,
			i.CreatedAt,
			i.UpdatedAt,
		)

	_, err = s.SQLExecutor.ExecWith(builder)
	if err != nil {
		return err
	}

	claims, err := json.Marshal(i.Claims)
	if err != nil {
		return err
	}

	q := s.SQLBuilder.Tenant().
		Insert(s.SQLBuilder.TableName("_auth_identity_login_id")).
		Columns(
			"id",
			"login_id_key",
			"login_id_type",
			"login_id",
			"original_login_id",
			"unique_key",
			"claims",
		).
		Values(
			i.ID,
			i.LoginIDKey,
			i.LoginIDType,
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
		Update(s.SQLBuilder.TableName("_auth_identity_login_id")).
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

	q = s.SQLBuilder.Tenant().
		Update(s.SQLBuilder.TableName("_auth_identity")).
		Set("updated_at", i.UpdatedAt).
		Where("id = ?", i.ID)

	_, err = s.SQLExecutor.ExecWith(q)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) Delete(i *Identity) error {
	q := s.SQLBuilder.Tenant().
		Delete(s.SQLBuilder.TableName("_auth_identity_login_id")).
		Where("id = ?", i.ID)

	_, err := s.SQLExecutor.ExecWith(q)
	if err != nil {
		return err
	}

	q = s.SQLBuilder.Tenant().
		Delete(s.SQLBuilder.TableName("_auth_identity")).
		Where("id = ?", i.ID)

	_, err = s.SQLExecutor.ExecWith(q)
	if err != nil {
		return err
	}

	return nil
}

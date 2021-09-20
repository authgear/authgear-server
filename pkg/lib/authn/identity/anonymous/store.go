package anonymous

import (
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/lib/pq"

	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
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
			"p.id",
			"p.labels",
			"p.user_id",
			"p.created_at",
			"p.updated_at",
			"a.key_id",
			"a.key",
		).
		From(s.SQLBuilder.TableName("_auth_identity"), "p").
		Join(s.SQLBuilder.TableName("_auth_identity_anonymous"), "a", "p.id = a.id")
}

func (s *Store) scan(scn db.Scanner) (*Identity, error) {
	i := &Identity{}
	var labels []byte

	err := scn.Scan(
		&i.ID,
		&labels,
		&i.UserID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.KeyID,
		&i.Key,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, identity.ErrIdentityNotFound
	} else if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(labels, &i.Labels); err != nil {
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
	labels, err := json.Marshal(i.Labels)
	if err != nil {
		return err
	}

	builder := s.SQLBuilder.
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
			authn.IdentityTypeAnonymous,
			i.UserID,
			i.CreatedAt,
			i.UpdatedAt,
		)

	_, err = s.SQLExecutor.ExecWith(builder)
	if err != nil {
		return err
	}

	q := s.SQLBuilder.
		Insert(s.SQLBuilder.TableName("_auth_identity_anonymous")).
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
	q := s.SQLBuilder.
		Delete(s.SQLBuilder.TableName("_auth_identity_anonymous")).
		Where("id = ?", i.ID)

	_, err := s.SQLExecutor.ExecWith(q)
	if err != nil {
		return err
	}

	q = s.SQLBuilder.
		Delete(s.SQLBuilder.TableName("_auth_identity")).
		Where("id = ?", i.ID)

	_, err = s.SQLExecutor.ExecWith(q)
	if err != nil {
		return err
	}

	return nil
}

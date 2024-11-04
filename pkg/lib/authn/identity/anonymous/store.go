package anonymous

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/model"
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
			"p.user_id",
			"p.created_at",
			"p.updated_at",
			"a.key_id",
			"a.key",
		).
		From(s.SQLBuilder.TableName("_auth_identity"), "p").
		Join(s.SQLBuilder.TableName("_auth_identity_anonymous"), "a", "p.id = a.id")
}

func (s *Store) scan(scn db.Scanner) (*identity.Anonymous, error) {
	i := &identity.Anonymous{}

	var keyID sql.NullString
	var key sql.NullString

	err := scn.Scan(
		&i.ID,
		&i.UserID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&keyID,
		&key,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, api.ErrIdentityNotFound
	} else if err != nil {
		return nil, err
	}

	i.KeyID = keyID.String
	i.Key = []byte(key.String)

	return i, nil
}

func (s *Store) GetMany(ctx context.Context, ids []string) ([]*identity.Anonymous, error) {
	builder := s.selectQuery().Where("p.id = ANY (?)", pq.Array(ids))

	rows, err := s.SQLExecutor.QueryWith(ctx, builder)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var is []*identity.Anonymous
	for rows.Next() {
		i, err := s.scan(rows)
		if err != nil {
			return nil, err
		}
		is = append(is, i)
	}

	return is, nil
}

func (s *Store) List(ctx context.Context, userID string) ([]*identity.Anonymous, error) {
	q := s.selectQuery().Where("p.user_id = ?", userID)

	rows, err := s.SQLExecutor.QueryWith(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var is []*identity.Anonymous
	for rows.Next() {
		i, err := s.scan(rows)
		if err != nil {
			return nil, err
		}
		is = append(is, i)
	}

	return is, nil
}

func (s *Store) Get(ctx context.Context, userID, id string) (*identity.Anonymous, error) {
	if userID == "" || id == "" {
		return nil, api.ErrIdentityNotFound
	}
	q := s.selectQuery().Where("p.user_id = ? AND p.id = ?", userID, id)
	rows, err := s.SQLExecutor.QueryRowWith(ctx, q)
	if err != nil {
		return nil, err
	}

	return s.scan(rows)
}

func (s *Store) GetByKeyID(ctx context.Context, keyID string) (*identity.Anonymous, error) {
	if keyID == "" {
		return nil, api.ErrIdentityNotFound
	}

	q := s.selectQuery().Where("a.key_id = ?", keyID)
	rows, err := s.SQLExecutor.QueryRowWith(ctx, q)
	if err != nil {
		return nil, err
	}

	return s.scan(rows)
}

func (s *Store) Create(ctx context.Context, i *identity.Anonymous) (err error) {
	builder := s.SQLBuilder.
		Insert(s.SQLBuilder.TableName("_auth_identity")).
		Columns(
			"id",
			"type",
			"user_id",
			"created_at",
			"updated_at",
		).
		Values(
			i.ID,
			model.IdentityTypeAnonymous,
			i.UserID,
			i.CreatedAt,
			i.UpdatedAt,
		)

	_, err = s.SQLExecutor.ExecWith(ctx, builder)
	if err != nil {
		return err
	}

	q := s.SQLBuilder.
		Insert(s.SQLBuilder.TableName("_auth_identity_anonymous"))

	if i.KeyID != "" {
		q = q.
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
	} else {
		q = q.
			Columns(
				"id",
			).
			Values(
				i.ID,
			)
	}

	_, err = s.SQLExecutor.ExecWith(ctx, q)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) Delete(ctx context.Context, i *identity.Anonymous) error {
	q := s.SQLBuilder.
		Delete(s.SQLBuilder.TableName("_auth_identity_anonymous")).
		Where("id = ?", i.ID)

	_, err := s.SQLExecutor.ExecWith(ctx, q)
	if err != nil {
		return err
	}

	q = s.SQLBuilder.
		Delete(s.SQLBuilder.TableName("_auth_identity")).
		Where("id = ?", i.ID)

	_, err = s.SQLExecutor.ExecWith(ctx, q)
	if err != nil {
		return err
	}

	return nil
}

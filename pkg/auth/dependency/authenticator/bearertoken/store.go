package bearertoken

import (
	"database/sql"
	"errors"
	"time"

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
	q1 := s.SQLBuilder.Tenant().
		Select(
			"a.id",
			"a.user_id",
			"abt.parent_id",
			"abt.token",
			"abt.created_at",
			"abt.expire_at",
		).
		From(s.SQLBuilder.FullTableName("authenticator"), "a").
		Join(
			s.SQLBuilder.FullTableName("authenticator_bearer_token"),
			"abt",
			"a.id = abt.id",
		).
		Where("a.user_id = ? AND a.id = ?", userID, id)

	row, err := s.SQLExecutor.QueryRowWith(q1)
	if err != nil {
		return nil, err
	}

	a := &Authenticator{}
	err = row.Scan(
		&a.ID,
		&a.UserID,
		&a.ParentID,
		&a.Token,
		&a.CreatedAt,
		&a.ExpireAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, authenticator.ErrAuthenticatorNotFound
	} else if err != nil {
		return nil, err
	}

	return a, nil
}

func (s *Store) GetByToken(userID string, token string) (*Authenticator, error) {
	q1 := s.SQLBuilder.Tenant().
		Select(
			"a.id",
			"a.user_id",
			"abt.parent_id",
			"abt.token",
			"abt.created_at",
			"abt.expire_at",
		).
		From(s.SQLBuilder.FullTableName("authenticator"), "a").
		Join(
			s.SQLBuilder.FullTableName("authenticator_bearer_token"),
			"abt",
			"a.id = abt.id",
		).
		// SECURITY(louis): Ideally we should compare the bearer token in constant time.
		// However, it requires us to fetch all bearer tokens. The number can be unbound
		// because we do not limit the number of the bearer tokens.
		Where("a.user_id = ? AND abt.token = ?", userID, token)

	row, err := s.SQLExecutor.QueryRowWith(q1)
	if err != nil {
		return nil, err
	}

	a := &Authenticator{}
	err = row.Scan(
		&a.ID,
		&a.UserID,
		&a.ParentID,
		&a.Token,
		&a.CreatedAt,
		&a.ExpireAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, authenticator.ErrAuthenticatorNotFound
	} else if err != nil {
		return nil, err
	}

	return a, nil
}

func (s *Store) List(userID string) ([]*Authenticator, error) {
	q1 := s.SQLBuilder.Tenant().
		Select(
			"a.id",
			"a.user_id",
			"abt.parent_id",
			"abt.token",
			"abt.created_at",
			"abt.expire_at",
		).
		From(s.SQLBuilder.FullTableName("authenticator"), "a").
		Join(
			s.SQLBuilder.FullTableName("authenticator_bearer_token"),
			"abt",
			"a.id = abt.id",
		).
		Where("a.user_id = ?", userID)

	rows, err := s.SQLExecutor.QueryWith(q1)
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
			&a.ParentID,
			&a.Token,
			&a.CreatedAt,
			&a.ExpireAt,
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
			Where("a.type = ? AND a.user_id = ?", authn.AuthenticatorTypeBearerToken, userID)

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
		Delete(s.SQLBuilder.FullTableName("authenticator_bearer_token")).
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

func (s *Store) DeleteAllByParentID(parentID string) error {
	ids, err := func() ([]string, error) {
		builder := s.SQLBuilder.Tenant().
			Select("a.id").
			From(s.SQLBuilder.FullTableName("authenticator"), "a").
			Join(
				s.SQLBuilder.FullTableName("authenticator_bearer_token"),
				"abt",
				"a.id = abt.id",
			).
			Where("abt.parent_id = ?", parentID)

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
		Delete(s.SQLBuilder.FullTableName("authenticator_bearer_token")).
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

func (s *Store) DeleteAllExpired(userID string, now time.Time) error {
	ids, err := func() ([]string, error) {
		builder := s.SQLBuilder.Tenant().
			Select("a.id").
			From(s.SQLBuilder.FullTableName("authenticator"), "a").
			Join(
				s.SQLBuilder.FullTableName("authenticator_bearer_token"),
				"abt",
				"a.id = abt.id",
			).
			Where("a.user_id = ? AND abt.expire_at < ?", userID, now)

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
		Delete(s.SQLBuilder.FullTableName("authenticator_bearer_token")).
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

func (s *Store) Create(a *Authenticator) error {
	q1 := s.SQLBuilder.Tenant().
		Insert(s.SQLBuilder.FullTableName("authenticator")).
		Columns(
			"id",
			"type",
			"user_id",
		).
		Values(
			a.ID,
			authn.AuthenticatorTypeBearerToken,
			a.UserID,
		)
	_, err := s.SQLExecutor.ExecWith(q1)
	if err != nil {
		return err
	}

	q2 := s.SQLBuilder.Tenant().
		Insert(s.SQLBuilder.FullTableName("authenticator_bearer_token")).
		Columns(
			"id",
			"parent_id",
			"token",
			"created_at",
			"expire_at",
		).
		Values(
			a.ID,
			a.ParentID,
			a.Token,
			a.CreatedAt,
			a.ExpireAt,
		)
	_, err = s.SQLExecutor.ExecWith(q2)
	if err != nil {
		return err
	}

	return nil
}

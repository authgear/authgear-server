package totp

import (
	"context"
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
			"at.secret",
			"at.display_name",
		).
		From(s.SQLBuilder.TableName("_auth_authenticator"), "a").
		Join(
			s.SQLBuilder.TableName("_auth_authenticator_totp"),
			"at",
			"a.id = at.id",
		)
}

func (s *Store) scan(scn db.Scanner) (*authenticator.TOTP, error) {
	a := &authenticator.TOTP{}

	err := scn.Scan(
		&a.ID,
		&a.UserID,
		&a.CreatedAt,
		&a.UpdatedAt,
		&a.IsDefault,
		&a.Kind,
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

func (s *Store) GetMany(ctx context.Context, ids []string) ([]*authenticator.TOTP, error) {
	builder := s.selectQuery().Where("a.id = ANY (?)", pq.Array(ids))

	rows, err := s.SQLExecutor.QueryWith(ctx, builder)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var as []*authenticator.TOTP
	for rows.Next() {
		a, err := s.scan(rows)
		if err != nil {
			return nil, err
		}
		as = append(as, a)
	}

	return as, nil
}

func (s *Store) Get(ctx context.Context, userID string, id string) (*authenticator.TOTP, error) {
	q := s.selectQuery().Where("a.user_id = ? AND a.id = ?", userID, id)

	row, err := s.SQLExecutor.QueryRowWith(ctx, q)
	if err != nil {
		return nil, err
	}

	return s.scan(row)
}

func (s *Store) List(ctx context.Context, userID string) ([]*authenticator.TOTP, error) {
	q := s.selectQuery().Where("a.user_id = ?", userID)

	rows, err := s.SQLExecutor.QueryWith(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var authenticators []*authenticator.TOTP
	for rows.Next() {
		a, err := s.scan(rows)
		if err != nil {
			return nil, err
		}
		authenticators = append(authenticators, a)
	}

	return authenticators, nil
}

func (s *Store) Delete(ctx context.Context, id string) error {
	q := s.SQLBuilder.
		Delete(s.SQLBuilder.TableName("_auth_authenticator_totp")).
		Where("id = ?", id)
	_, err := s.SQLExecutor.ExecWith(ctx, q)
	if err != nil {
		return err
	}

	q = s.SQLBuilder.
		Delete(s.SQLBuilder.TableName("_auth_authenticator")).
		Where("id = ?", id)
	_, err = s.SQLExecutor.ExecWith(ctx, q)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) Create(ctx context.Context, a *authenticator.TOTP) (err error) {
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
			model.AuthenticatorTypeTOTP,
			a.UserID,
			a.CreatedAt,
			a.UpdatedAt,
			a.IsDefault,
			a.Kind,
		)
	_, err = s.SQLExecutor.ExecWith(ctx, q)
	if err != nil {
		return err
	}

	q = s.SQLBuilder.
		Insert(s.SQLBuilder.TableName("_auth_authenticator_totp")).
		Columns(
			"id",
			"secret",
			"display_name",
		).
		Values(
			a.ID,
			a.Secret,
			a.DisplayName,
		)
	_, err = s.SQLExecutor.ExecWith(ctx, q)
	if err != nil {
		return err
	}

	return nil
}

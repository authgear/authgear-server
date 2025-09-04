package oob

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

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
			"a.type",
			"a.user_id",
			"a.created_at",
			"a.updated_at",
			"a.is_default",
			"a.kind",
			"ao.phone",
			"ao.email",
			"ao.preferred_channel",
		).
		From(s.SQLBuilder.TableName("_auth_authenticator"), "a").
		Join(s.SQLBuilder.TableName("_auth_authenticator_oob"), "ao", "a.id = ao.id")
}

func (s *Store) scan(scn db.Scanner) (*authenticator.OOBOTP, error) {
	a := &authenticator.OOBOTP{}

	err := scn.Scan(
		&a.ID,
		&a.OOBAuthenticatorType,
		&a.UserID,
		&a.CreatedAt,
		&a.UpdatedAt,
		&a.IsDefault,
		&a.Kind,
		&a.Phone,
		&a.Email,
		&a.PreferredChannel,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, authenticator.ErrAuthenticatorNotFound
	} else if err != nil {
		return nil, err
	}

	return a, nil
}

func (s *Store) Get(ctx context.Context, userID string, id string) (*authenticator.OOBOTP, error) {
	q := s.selectQuery().Where("a.user_id = ? AND a.id = ?", userID, id)

	row, err := s.SQLExecutor.QueryRowWith(ctx, q)
	if err != nil {
		return nil, err
	}

	return s.scan(row)
}

func (s *Store) GetMany(ctx context.Context, ids []string) ([]*authenticator.OOBOTP, error) {
	builder := s.selectQuery().Where("a.id = ANY (?)", pq.Array(ids))

	rows, err := s.SQLExecutor.QueryWith(ctx, builder)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var as []*authenticator.OOBOTP
	for rows.Next() {
		a, err := s.scan(rows)
		if err != nil {
			return nil, err
		}
		as = append(as, a)
	}

	return as, nil
}

func (s *Store) List(ctx context.Context, userID string) ([]*authenticator.OOBOTP, error) {
	q := s.selectQuery().Where("a.user_id = ?", userID)

	rows, err := s.SQLExecutor.QueryWith(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var authenticators []*authenticator.OOBOTP
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
		Delete(s.SQLBuilder.TableName("_auth_authenticator_oob")).
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

func (s *Store) Create(ctx context.Context, a *authenticator.OOBOTP) (err error) {
	if a.OOBAuthenticatorType != model.AuthenticatorTypeOOBEmail &&
		a.OOBAuthenticatorType != model.AuthenticatorTypeOOBSMS {
		return errors.New("invalid oob authenticator type")
	}

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
			a.OOBAuthenticatorType,
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
		Insert(s.SQLBuilder.TableName("_auth_authenticator_oob")).
		Columns(
			"id",
			"phone",
			"email",
			"preferred_channel",
		).
		Values(
			a.ID,
			a.Phone,
			a.Email,
			a.PreferredChannel,
		)
	_, err = s.SQLExecutor.ExecWith(ctx, q)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) Update(ctx context.Context, a *authenticator.OOBOTP) (err error) {
	if a.OOBAuthenticatorType != model.AuthenticatorTypeOOBEmail &&
		a.OOBAuthenticatorType != model.AuthenticatorTypeOOBSMS {
		return errors.New("invalid oob authenticator type")
	}

	q := s.SQLBuilder.
		Update(s.SQLBuilder.TableName("_auth_authenticator")).
		Set("updated_at", a.UpdatedAt).
		Where("id = ?", a.ID)
	result, err := s.SQLExecutor.ExecWith(ctx, q)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return authenticator.ErrAuthenticatorNotFound
	} else if rowsAffected > 1 {
		panic(fmt.Sprintf("authenticator_oob: want 1 row updated, got %v", rowsAffected))
	}

	q = s.SQLBuilder.
		Update(s.SQLBuilder.TableName("_auth_authenticator_oob")).
		Set("phone", a.Phone).
		Set("email", a.Email).
		Set("preferred_channel", a.PreferredChannel).
		Where("id = ?", a.ID)
	result, err = s.SQLExecutor.ExecWith(ctx, q)
	if err != nil {
		return err
	}
	rowsAffected, err = result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return authenticator.ErrAuthenticatorNotFound
	} else if rowsAffected > 1 {
		panic(fmt.Sprintf("authenticator_oob: want 1 row updated, got %v", rowsAffected))
	}

	return nil
}

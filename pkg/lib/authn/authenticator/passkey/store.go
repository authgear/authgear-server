package passkey

import (
	"context"
	"database/sql"
	"encoding/json"
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
			"ap.credential_id",
			"ap.creation_options",
			"ap.attestation_response",
			"ap.sign_count",
		).
		From(s.SQLBuilder.TableName("_auth_authenticator"), "a").
		Join(s.SQLBuilder.TableName("_auth_authenticator_passkey"), "ap", "a.id = ap.id")
}

func (s *Store) scan(scanner db.Scanner) (*authenticator.Passkey, error) {
	a := &authenticator.Passkey{}
	var creationOptionsBytes []byte

	err := scanner.Scan(
		&a.ID,
		&a.UserID,
		&a.CreatedAt,
		&a.UpdatedAt,
		&a.IsDefault,
		&a.Kind,
		&a.CredentialID,
		&creationOptionsBytes,
		&a.AttestationResponse,
		&a.SignCount,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, authenticator.ErrAuthenticatorNotFound
	} else if err != nil {
		return nil, err
	}

	err = json.Unmarshal(creationOptionsBytes, &a.CreationOptions)
	if err != nil {
		return nil, err
	}

	return a, nil
}

func (s *Store) GetMany(ctx context.Context, ids []string) ([]*authenticator.Passkey, error) {
	builder := s.selectQuery().Where("a.id = ANY (?)", pq.Array(ids))

	rows, err := s.SQLExecutor.QueryWith(ctx, builder)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var as []*authenticator.Passkey
	for rows.Next() {
		a, err := s.scan(rows)
		if err != nil {
			return nil, err
		}
		as = append(as, a)
	}

	return as, nil
}

func (s *Store) Get(ctx context.Context, userID string, id string) (*authenticator.Passkey, error) {
	q := s.selectQuery().Where("a.user_id = ? AND a.id = ?", userID, id)

	row, err := s.SQLExecutor.QueryRowWith(ctx, q)
	if err != nil {
		return nil, err
	}

	return s.scan(row)
}

func (s *Store) List(ctx context.Context, userID string) ([]*authenticator.Passkey, error) {
	q := s.selectQuery().Where("a.user_id = ?", userID)

	rows, err := s.SQLExecutor.QueryWith(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var authenticators []*authenticator.Passkey
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
		Delete(s.SQLBuilder.TableName("_auth_authenticator_passkey")).
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

func (s *Store) Create(ctx context.Context, a *authenticator.Passkey) (err error) {
	creationOptionsBytes, err := json.Marshal(a.CreationOptions)
	if err != nil {
		return err
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
			model.AuthenticatorTypePasskey,
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
		Insert(s.SQLBuilder.TableName("_auth_authenticator_passkey")).
		Columns(
			"id",
			"credential_id",
			"creation_options",
			"attestation_response",
			"sign_count",
		).
		Values(
			a.ID,
			a.CredentialID,
			creationOptionsBytes,
			a.AttestationResponse,
			a.SignCount,
		)
	_, err = s.SQLExecutor.ExecWith(ctx, q)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) UpdateSignCount(ctx context.Context, a *authenticator.Passkey) error {
	q := s.SQLBuilder.
		Update(s.SQLBuilder.TableName("_auth_authenticator_passkey")).
		Set("sign_count", a.SignCount).
		Where("id = ?", a.ID)
	_, err := s.SQLExecutor.ExecWith(ctx, q)
	if err != nil {
		return err
	}

	q = s.SQLBuilder.
		Update(s.SQLBuilder.TableName("_auth_authenticator")).
		Set("updated_at", a.UpdatedAt).
		Where("id = ?", a.ID)

	_, err = s.SQLExecutor.ExecWith(ctx, q)
	if err != nil {
		return err
	}

	return nil
}

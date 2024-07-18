package passkey

import (
	"database/sql"
	"encoding/json"
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
			"i.id",
			"i.user_id",
			"i.created_at",
			"i.updated_at",
			"p.credential_id",
			"p.creation_options",
			"p.attestation_response",
		).
		From(s.SQLBuilder.TableName("_auth_identity"), "i").
		Join(s.SQLBuilder.TableName("_auth_identity_passkey"), "p", "i.id = p.id")
}

func (s *Store) scan(scanner db.Scanner) (*identity.Passkey, error) {
	i := &identity.Passkey{}
	var creationOptionsBytes []byte

	err := scanner.Scan(
		&i.ID,
		&i.UserID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.CredentialID,
		&creationOptionsBytes,
		&i.AttestationResponse,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, api.ErrIdentityNotFound
	} else if err != nil {
		return nil, err
	}

	err = json.Unmarshal(creationOptionsBytes, &i.CreationOptions)
	if err != nil {
		return nil, err
	}

	return i, nil
}

func (s *Store) GetMany(ids []string) ([]*identity.Passkey, error) {
	builder := s.selectQuery().Where("i.id = ANY (?)", pq.Array(ids))

	rows, err := s.SQLExecutor.QueryWith(builder)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var is []*identity.Passkey
	for rows.Next() {
		i, err := s.scan(rows)
		if err != nil {
			return nil, err
		}
		is = append(is, i)
	}

	return is, nil
}

func (s *Store) List(userID string) ([]*identity.Passkey, error) {
	q := s.selectQuery().Where("i.user_id = ?", userID)

	rows, err := s.SQLExecutor.QueryWith(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var is []*identity.Passkey
	for rows.Next() {
		i, err := s.scan(rows)
		if err != nil {
			return nil, err
		}
		is = append(is, i)
	}

	return is, nil
}

func (s *Store) Get(userID, id string) (*identity.Passkey, error) {
	q := s.selectQuery().Where("i.user_id = ? AND i.id = ?", userID, id)
	rows, err := s.SQLExecutor.QueryRowWith(q)
	if err != nil {
		return nil, err
	}

	return s.scan(rows)
}

func (s *Store) GetByCredentialID(credentialID string) (*identity.Passkey, error) {
	q := s.selectQuery().Where("p.credential_id = ?", credentialID)
	rows, err := s.SQLExecutor.QueryRowWith(q)
	if err != nil {
		return nil, err
	}

	return s.scan(rows)
}

func (s *Store) Create(i *identity.Passkey) error {
	creationOptionsBytes, err := json.Marshal(i.CreationOptions)
	if err != nil {
		return err
	}

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
			model.IdentityTypePasskey,
			i.UserID,
			i.CreatedAt,
			i.UpdatedAt,
		)

	_, err = s.SQLExecutor.ExecWith(builder)
	if err != nil {
		return err
	}

	q := s.SQLBuilder.
		Insert(s.SQLBuilder.TableName("_auth_identity_passkey")).
		Columns(
			"id",
			"credential_id",
			"creation_options",
			"attestation_response",
		).
		Values(
			i.ID,
			i.CredentialID,
			creationOptionsBytes,
			i.AttestationResponse,
		)

	_, err = s.SQLExecutor.ExecWith(q)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) Delete(i *identity.Passkey) error {
	q := s.SQLBuilder.
		Delete(s.SQLBuilder.TableName("_auth_identity_passkey")).
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

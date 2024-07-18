package siwe

import (
	"database/sql"
	"errors"

	"github.com/goccy/go-json"
	"github.com/lib/pq"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/util/web3"
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
			"s.chain_id",
			"s.address",
			"s.data",
		).
		From(s.SQLBuilder.TableName("_auth_identity"), "i").
		Join(s.SQLBuilder.TableName("_auth_identity_siwe"), "s", "i.id = s.id")
}

func (s *Store) scan(scanner db.Scanner) (*identity.SIWE, error) {
	i := &identity.SIWE{}
	var address string
	var data []byte
	err := scanner.Scan(
		&i.ID,
		&i.UserID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.ChainID,
		&address,
		&data,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, api.ErrIdentityNotFound
	} else if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(data, &i.Data); err != nil {
		return nil, err
	}

	encodedAddress, err := web3.NewEIP55(address)
	if err != nil {
		return nil, err
	}

	i.Address = encodedAddress

	return i, nil
}

func (s *Store) GetMany(ids []string) ([]*identity.SIWE, error) {
	builder := s.selectQuery().Where("i.id = ANY (?)", pq.Array(ids))

	rows, err := s.SQLExecutor.QueryWith(builder)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var is []*identity.SIWE
	for rows.Next() {
		i, err := s.scan(rows)
		if err != nil {
			return nil, err
		}
		is = append(is, i)
	}

	return is, nil
}

func (s *Store) List(userID string) ([]*identity.SIWE, error) {
	q := s.selectQuery().Where("i.user_id = ?", userID)

	rows, err := s.SQLExecutor.QueryWith(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var is []*identity.SIWE
	for rows.Next() {
		i, err := s.scan(rows)
		if err != nil {
			return nil, err
		}
		is = append(is, i)
	}

	return is, nil
}

func (s *Store) Get(userID, id string) (*identity.SIWE, error) {
	q := s.selectQuery().Where("i.user_id = ? AND i.id = ?", userID, id)
	rows, err := s.SQLExecutor.QueryRowWith(q)
	if err != nil {
		return nil, err
	}

	return s.scan(rows)
}

func (s *Store) GetByAddress(chainID int, address web3.EIP55) (*identity.SIWE, error) {
	addrHex, err := address.ToHexstring()
	if err != nil {
		return nil, err
	}
	q := s.selectQuery().Where("s.chain_id = ? AND s.address = ?", chainID, addrHex.String())
	rows, err := s.SQLExecutor.QueryRowWith(q)
	if err != nil {
		return nil, err
	}
	return s.scan(rows)
}

func (s *Store) Create(i *identity.SIWE) error {

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
			model.IdentityTypeSIWE,
			i.UserID,
			i.CreatedAt,
			i.UpdatedAt,
		)

	_, err := s.SQLExecutor.ExecWith(builder)
	if err != nil {
		return err
	}

	data, err := json.Marshal(i.Data)
	if err != nil {
		return err
	}

	address, err := i.Address.ToHexstring()
	if err != nil {
		return err
	}

	q := s.SQLBuilder.
		Insert(s.SQLBuilder.TableName("_auth_identity_siwe")).
		Columns(
			"id",
			"address",
			"chain_id",
			"data",
		).
		Values(
			i.ID,
			address,
			i.ChainID,
			data,
		)

	_, err = s.SQLExecutor.ExecWith(q)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) Delete(i *identity.SIWE) error {
	q := s.SQLBuilder.
		Delete(s.SQLBuilder.TableName("_auth_identity_siwe")).
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

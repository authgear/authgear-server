package ldap

import (
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/lib/pq"

	"github.com/authgear/authgear-server/pkg/api"
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

			"l.server_name",
			"l.user_id_attribute",
			"l.user_id_attribute_value",
			"l.claims",
			"l.raw_entry_json",
		).
		From(s.SQLBuilder.TableName("_auth_identity"), "p").
		Join(s.SQLBuilder.TableName("_auth_identity_ldap"), "l", "p.id = l.id")
}

func (s *Store) scan(scn db.Scanner) (*identity.LDAP, error) {
	i := &identity.LDAP{}
	var claims []byte
	var rawEntryJSON []byte

	err := scn.Scan(
		&i.ID,
		&i.UserID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.ServerName,
		&i.UserIDAttribute,
		&i.UserIDAttributeValue,
		&claims,
		&rawEntryJSON,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, api.ErrIdentityNotFound
	} else if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(claims, &i.Claims); err != nil {
		return nil, err
	}
	if err = json.Unmarshal(rawEntryJSON, &i.RawEntryJSON); err != nil {
		return nil, err
	}

	return i, nil
}

func (s *Store) Get(userID string, id string) (*identity.LDAP, error) {
	q := s.selectQuery().Where("p.user_id = ? AND p.id = ?", userID, id)
	rows, err := s.SQLExecutor.QueryRowWith(q)
	if err != nil {
		return nil, err
	}
	return s.scan(rows)
}

func (s *Store) GetMany(ids []string) ([]*identity.LDAP, error) {
	builder := s.selectQuery().Where("p.id = ANY (?)", pq.Array(ids))

	rows, err := s.SQLExecutor.QueryWith(builder)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var is []*identity.LDAP
	for rows.Next() {
		i, err := s.scan(rows)
		if err != nil {
			return nil, err
		}
		is = append(is, i)
	}

	return is, nil
}

func (s *Store) List(userID string) ([]*identity.LDAP, error) {
	builder := s.selectQuery().Where("p.user_id = ANY (?)", pq.Array(userID))

	rows, err := s.SQLExecutor.QueryWith(builder)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var is []*identity.LDAP
	for rows.Next() {
		i, err := s.scan(rows)
		if err != nil {
			return nil, err
		}
		is = append(is, i)
	}

	return is, nil
}

func (s *Store) GetByServerUserID(serverName string, userIDAttribute string, userIDAttributeValue string) (*identity.LDAP, error) {
	q := s.selectQuery().
		Where(
			"o.server_name = ? AND o.user_id_attribute = ? AND o.user_id_attribute_value = ?",
			serverName,
			userIDAttribute,
			userIDAttributeValue,
		)
	rows, err := s.SQLExecutor.QueryRowWith(q)
	if err != nil {
		return nil, err
	}
	return s.scan(rows)
}

package oauth

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"
	"github.com/lib/pq"

	"github.com/authgear/oauthrelyingparty/pkg/api/oauthrelyingparty"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
)

type Store struct {
	SQLBuilder     *appdb.SQLBuilderApp
	SQLExecutor    *appdb.SQLExecutor
	IdentityConfig *config.IdentityConfig
}

func (s *Store) selectQuery() db.SelectBuilder {
	return s.SQLBuilder.
		Select(
			"p.id",
			"p.user_id",
			"p.created_at",
			"p.updated_at",
			"o.provider_type",
			"o.provider_keys",
			"o.provider_user_id",
			"o.profile",
			"o.claims",
		).
		From(s.SQLBuilder.TableName("_auth_identity"), "p").
		Join(s.SQLBuilder.TableName("_auth_identity_oauth"), "o", "p.id = o.id")
}

func (s *Store) scan(scn db.Scanner) (*identity.OAuth, error) {
	i := &identity.OAuth{}
	var providerKeys []byte
	var profile []byte
	var claims []byte

	err := scn.Scan(
		&i.ID,
		&i.UserID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.ProviderID.Type,
		&providerKeys,
		&i.ProviderSubjectID,
		&profile,
		&claims,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, identity.ErrIdentityNotFound
	} else if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(providerKeys, &i.ProviderID.Keys); err != nil {
		return nil, err
	}
	if err = json.Unmarshal(profile, &i.UserProfile); err != nil {
		return nil, err
	}
	if err = json.Unmarshal(claims, &i.Claims); err != nil {
		return nil, err
	}

	alias := ""
	for _, providerConfig := range s.IdentityConfig.OAuth.Providers {
		providerID := providerConfig.AsProviderConfig().ProviderID()
		if providerID.Equal(i.ProviderID) {
			alias = providerConfig.Alias()
		}
	}
	if alias != "" {
		i.ProviderAlias = alias
	}

	return i, nil
}

func (s *Store) GetMany(ids []string) ([]*identity.OAuth, error) {
	builder := s.selectQuery().Where("p.id = ANY (?)", pq.Array(ids))

	rows, err := s.SQLExecutor.QueryWith(builder)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var is []*identity.OAuth
	for rows.Next() {
		i, err := s.scan(rows)
		if err != nil {
			return nil, err
		}
		is = append(is, i)
	}

	return is, nil
}

func (s *Store) List(userID string) ([]*identity.OAuth, error) {
	q := s.selectQuery().Where("p.user_id = ?", userID)

	rows, err := s.SQLExecutor.QueryWith(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var is []*identity.OAuth
	for rows.Next() {
		i, err := s.scan(rows)
		if err != nil {
			return nil, err
		}
		is = append(is, i)
	}

	return is, nil
}

func (s *Store) ListByClaim(name string, value string) ([]*identity.OAuth, error) {
	q := s.selectQuery().
		Where("(o.claims ->> ?) = ?", name, value)

	rows, err := s.SQLExecutor.QueryWith(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var is []*identity.OAuth
	for rows.Next() {
		i, err := s.scan(rows)
		if err != nil {
			return nil, err
		}
		is = append(is, i)
	}

	return is, nil
}

func (s *Store) ListByClaimJSONPointer(pointer jsonpointer.T, value string) ([]*identity.OAuth, error) {
	q := s.selectQuery().
		Where("(o.claims #>> ?) = ?", pq.Array(pointer), value)

	rows, err := s.SQLExecutor.QueryWith(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var is []*identity.OAuth
	for rows.Next() {
		i, err := s.scan(rows)
		if err != nil {
			return nil, err
		}
		is = append(is, i)
	}

	return is, nil
}

func (s *Store) Get(userID string, id string) (*identity.OAuth, error) {
	q := s.selectQuery().Where("p.user_id = ? AND p.id = ?", userID, id)
	rows, err := s.SQLExecutor.QueryRowWith(q)
	if err != nil {
		return nil, err
	}

	return s.scan(rows)
}

func (s *Store) GetByProviderSubject(providerID oauthrelyingparty.ProviderID, subjectID string) (*identity.OAuth, error) {
	providerKeys, err := json.Marshal(providerID.Keys)
	if err != nil {
		return nil, err
	}

	q := s.selectQuery().Where(
		"o.provider_type = ? AND o.provider_keys = ? AND o.provider_user_id = ?",
		providerID.Type, providerKeys, subjectID)
	rows, err := s.SQLExecutor.QueryRowWith(q)
	if err != nil {
		return nil, err
	}

	return s.scan(rows)
}

func (s *Store) GetByUserProvider(userID string, providerID oauthrelyingparty.ProviderID) (*identity.OAuth, error) {
	providerKeys, err := json.Marshal(providerID.Keys)
	if err != nil {
		return nil, err
	}

	q := s.selectQuery().Where(
		"o.provider_type = ? AND o.provider_keys = ? AND p.user_id = ?",
		providerID.Type, providerKeys, userID)
	rows, err := s.SQLExecutor.QueryRowWith(q)
	if err != nil {
		return nil, err
	}

	return s.scan(rows)
}

func (s *Store) Create(i *identity.OAuth) (err error) {
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
			model.IdentityTypeOAuth,
			i.UserID,
			i.CreatedAt,
			i.UpdatedAt,
		)

	_, err = s.SQLExecutor.ExecWith(builder)
	if err != nil {
		return err
	}

	providerKeys, err := json.Marshal(i.ProviderID.Keys)
	if err != nil {
		return err
	}
	profile, err := json.Marshal(i.UserProfile)
	if err != nil {
		return err
	}
	claims, err := json.Marshal(i.Claims)
	if err != nil {
		return err
	}

	q := s.SQLBuilder.
		Insert(s.SQLBuilder.TableName("_auth_identity_oauth")).
		Columns(
			"id",
			"provider_type",
			"provider_keys",
			"provider_user_id",
			"profile",
			"claims",
		).
		Values(
			i.ID,
			i.ProviderID.Type,
			providerKeys,
			i.ProviderSubjectID,
			profile,
			claims,
		)

	_, err = s.SQLExecutor.ExecWith(q)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) Update(i *identity.OAuth) error {
	profile, err := json.Marshal(i.UserProfile)
	if err != nil {
		return err
	}
	claims, err := json.Marshal(i.Claims)
	if err != nil {
		return err
	}

	q := s.SQLBuilder.
		Update(s.SQLBuilder.TableName("_auth_identity_oauth")).
		Set("claims", claims).
		Set("profile", profile).
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

	q = s.SQLBuilder.
		Update(s.SQLBuilder.TableName("_auth_identity")).
		Set("updated_at", i.UpdatedAt).
		Where("id = ?", i.ID)

	_, err = s.SQLExecutor.ExecWith(q)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) Delete(i *identity.OAuth) error {
	q := s.SQLBuilder.
		Delete(s.SQLBuilder.TableName("_auth_identity_oauth")).
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

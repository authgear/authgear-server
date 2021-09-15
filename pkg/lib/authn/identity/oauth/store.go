package oauth

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/lib/pq"

	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
)

type Store struct {
	SQLBuilder  *appdb.SQLBuilder
	SQLExecutor *appdb.SQLExecutor
}

func (s *Store) selectQuery() db.SelectBuilder {
	return s.SQLBuilder.Tenant().
		Select(
			"p.id",
			"p.labels",
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

func (s *Store) scan(scn db.Scanner) (*Identity, error) {
	i := &Identity{}
	var labels []byte
	var providerKeys []byte
	var profile []byte
	var claims []byte

	err := scn.Scan(
		&i.ID,
		&labels,
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

	if err = json.Unmarshal(labels, &i.Labels); err != nil {
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
	q := s.selectQuery().
		Where("(o.claims #>> ?) = ?", pq.Array([]string{name}), value)

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

func (s *Store) Get(userID string, id string) (*Identity, error) {
	q := s.selectQuery().Where("p.user_id = ? AND p.id = ?", userID, id)
	rows, err := s.SQLExecutor.QueryRowWith(q)
	if err != nil {
		return nil, err
	}

	return s.scan(rows)
}

func (s *Store) GetByProviderSubject(provider config.ProviderID, subjectID string) (*Identity, error) {
	providerKeys, err := json.Marshal(provider.Keys)
	if err != nil {
		return nil, err
	}

	q := s.selectQuery().Where(
		"o.provider_type = ? AND o.provider_keys = ? AND o.provider_user_id = ?",
		provider.Type, providerKeys, subjectID)
	rows, err := s.SQLExecutor.QueryRowWith(q)
	if err != nil {
		return nil, err
	}

	return s.scan(rows)
}

func (s *Store) GetByUserProvider(userID string, provider config.ProviderID) (*Identity, error) {
	providerKeys, err := json.Marshal(provider.Keys)
	if err != nil {
		return nil, err
	}

	q := s.selectQuery().Where(
		"o.provider_type = ? AND o.provider_keys = ? AND p.user_id = ?",
		provider.Type, providerKeys, userID)
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

	builder := s.SQLBuilder.Tenant().
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
			authn.IdentityTypeOAuth,
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

	q := s.SQLBuilder.Tenant().
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

func (s *Store) Update(i *Identity) error {
	profile, err := json.Marshal(i.UserProfile)
	if err != nil {
		return err
	}
	claims, err := json.Marshal(i.Claims)
	if err != nil {
		return err
	}

	q := s.SQLBuilder.Tenant().
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

	q = s.SQLBuilder.Tenant().
		Update(s.SQLBuilder.TableName("_auth_identity")).
		Set("updated_at", i.UpdatedAt).
		Where("id = ?", i.ID)

	_, err = s.SQLExecutor.ExecWith(q)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) Delete(i *Identity) error {
	q := s.SQLBuilder.Tenant().
		Delete(s.SQLBuilder.TableName("_auth_identity_oauth")).
		Where("id = ?", i.ID)

	_, err := s.SQLExecutor.ExecWith(q)
	if err != nil {
		return err
	}

	q = s.SQLBuilder.Tenant().
		Delete(s.SQLBuilder.TableName("_auth_identity")).
		Where("id = ?", i.ID)

	_, err = s.SQLExecutor.ExecWith(q)
	if err != nil {
		return err
	}

	return nil
}

package pq

import (
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/jmoiron/sqlx"

	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
)

type AuthorizationStore struct {
	SQLBuilder  db.SQLBuilder
	SQLExecutor db.SQLExecutor
}

func (s *AuthorizationStore) selectQuery() db.SelectBuilder {
	return s.SQLBuilder.Tenant().Select(
		"id",
		"labels",
		"app_id",
		"client_id",
		"user_id",
		"created_at",
		"updated_at",
		"scopes",
	).
		From(s.SQLBuilder.FullTableName("oauth_authorization"))
}

func (s *AuthorizationStore) Get(userID, clientID string) (*oauth.Authorization, error) {
	builder := s.selectQuery().
		Where("user_id = ? AND client_id = ?", userID, clientID)

	scanner, err := s.SQLExecutor.QueryRowWith(builder)
	if err != nil {
		return nil, err
	}

	return s.scanAuthz(scanner)
}

func (s *AuthorizationStore) GetByID(id string) (*oauth.Authorization, error) {
	builder := s.selectQuery().
		Where("id = ?", id)

	scanner, err := s.SQLExecutor.QueryRowWith(builder)
	if err != nil {
		return nil, err
	}

	return s.scanAuthz(scanner)
}

func (s *AuthorizationStore) scanAuthz(scn sqlx.ColScanner) (*oauth.Authorization, error) {
	authz := &oauth.Authorization{}

	var labels []byte
	var scopeBytes []byte

	err := scn.Scan(
		&authz.ID,
		&labels,
		&authz.AppID,
		&authz.ClientID,
		&authz.UserID,
		&authz.CreatedAt,
		&authz.UpdatedAt,
		&scopeBytes,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, oauth.ErrAuthorizationNotFound
	} else if err != nil {
		return nil, err
	}

	err = json.Unmarshal(labels, &authz.Labels)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(scopeBytes, &authz.Scopes)
	if err != nil {
		return nil, err
	}

	return authz, nil
}

func (s *AuthorizationStore) Create(authz *oauth.Authorization) error {
	labels, err := json.Marshal(authz.Labels)
	if err != nil {
		return err
	}

	scopeBytes, err := json.Marshal(authz.Scopes)
	if err != nil {
		return err
	}

	builder := s.SQLBuilder.Tenant().
		Insert(s.SQLBuilder.FullTableName("oauth_authorization")).
		Columns(
			"id",
			"labels",
			"client_id",
			"user_id",
			"created_at",
			"updated_at",
			"scopes",
		).
		Values(
			authz.ID,
			labels,
			authz.ClientID,
			authz.UserID,
			authz.CreatedAt,
			authz.UpdatedAt,
			scopeBytes,
		)

	_, err = s.SQLExecutor.ExecWith(builder)
	if err != nil {
		return err
	}

	return nil
}

func (s *AuthorizationStore) Delete(authz *oauth.Authorization) error {
	builder := s.SQLBuilder.Tenant().
		Delete(s.SQLBuilder.FullTableName("oauth_authorization")).
		Where("id = ?", authz.ID)

	_, err := s.SQLExecutor.ExecWith(builder)
	if err != nil {
		return err
	}

	return nil
}

func (s *AuthorizationStore) ResetAll(userID string) error {
	builder := s.SQLBuilder.Tenant().
		Delete(s.SQLBuilder.FullTableName("oauth_authorization")).
		Where("user_id = ?", userID)

	_, err := s.SQLExecutor.ExecWith(builder)
	if err != nil {
		return err
	}

	return nil
}

func (s *AuthorizationStore) UpdateScopes(authz *oauth.Authorization) error {
	scopeBytes, err := json.Marshal(authz.Scopes)
	if err != nil {
		return err
	}

	builder := s.SQLBuilder.Tenant().
		Update(s.SQLBuilder.FullTableName("oauth_authorization")).
		Set("updated_at", authz.UpdatedAt).
		Set("scopes", scopeBytes).
		Where("id = ?", authz.ID)

	_, err = s.SQLExecutor.ExecWith(builder)
	if err != nil {
		return err
	}

	return nil
}

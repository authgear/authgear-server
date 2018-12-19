package oauth

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
)

type providerImpl struct {
	sqlBuilder  db.SQLBuilder
	sqlExecutor db.SQLExecutor
	logger      *logrus.Entry
}

func newProvider(
	builder db.SQLBuilder,
	executor db.SQLExecutor,
	logger *logrus.Entry,
) *providerImpl {
	return &providerImpl{
		sqlBuilder:  builder,
		sqlExecutor: executor,
		logger:      logger,
	}
}

func NewProvider(
	builder db.SQLBuilder,
	executor db.SQLExecutor,
	logger *logrus.Entry,
) Provider {
	return newProvider(builder, executor, logger)
}

func (p providerImpl) GetPrincipalByProviderUserID(providerName string, providerUserID string) (*Principal, error) {
	principal := Principal{}
	principal.ProviderName = providerName
	principal.ProviderUserID = providerUserID

	builder := p.sqlBuilder.Select("p.id", "p.user_id").
		From(fmt.Sprintf("%s as p", p.sqlBuilder.FullTableName("principal"))).
		Join(p.sqlBuilder.FullTableName("provider_oauth")+" AS oauth ON p.id = oauth.principal_id").
		Where("oauth.oauth_provider = ? AND oauth.provider_user_id = ? AND p.provider = 'oauth'", providerName, providerUserID)
	scanner := p.sqlExecutor.QueryRowWith(builder)

	err := scanner.Scan(
		&principal.ID,
		&principal.UserID,
	)

	if err == sql.ErrNoRows {
		err = skydb.ErrUserNotFound
	}

	if err != nil {
		return nil, err
	}

	return &principal, nil
}

func (p providerImpl) GetPrincipalByUserID(providerName string, userID string) (*Principal, error) {
	principal := Principal{}
	principal.UserID = userID

	builder := p.sqlBuilder.Select("p.id", "oauth.oauth_provider", "oauth.provider_user_id").
		From(fmt.Sprintf("%s as p", p.sqlBuilder.FullTableName("principal"))).
		Join(p.sqlBuilder.FullTableName("provider_oauth")+" AS oauth ON p.id = oauth.principal_id").
		Where("oauth.oauth_provider = ? AND p.user_id = ? AND p.provider = 'oauth'", providerName, userID)
	scanner := p.sqlExecutor.QueryRowWith(builder)

	err := scanner.Scan(
		&principal.ID,
		&principal.ProviderName,
		&principal.ProviderUserID,
	)

	if err == sql.ErrNoRows {
		err = skydb.ErrUserNotFound
	}

	if err != nil {
		return nil, err
	}

	return &principal, nil
}

func (p *providerImpl) CreatePrincipal(principal Principal) (err error) {
	var (
		createdAt *time.Time
		updatedAt *time.Time
	)
	createdAt = principal.CreatedAt
	if createdAt != nil && createdAt.IsZero() {
		createdAt = nil
	}
	updatedAt = principal.UpdatedAt
	if updatedAt != nil && updatedAt.IsZero() {
		updatedAt = nil
	}

	// Create principal
	builder := p.sqlBuilder.Insert(p.sqlBuilder.FullTableName("principal")).Columns(
		"id",
		"provider",
		"user_id",
	).Values(
		principal.ID,
		providerName,
		principal.UserID,
	)

	_, err = p.sqlExecutor.ExecWith(builder)
	if err != nil {
		return
	}

	var accessTokenRespBytes []byte
	accessTokenRespBytes, err = json.Marshal(principal.AccessTokenResp)
	if err != nil {
		return
	}

	var userProfileBytes []byte
	userProfileBytes, err = json.Marshal(principal.UserProfile)
	if err != nil {
		return
	}

	builder = p.sqlBuilder.Insert(p.sqlBuilder.FullTableName("provider_oauth")).Columns(
		"principal_id",
		"oauth_provider",
		"provider_user_id",
		"token_response",
		"profile",
		"_created_at",
		"_updated_at",
	).Values(
		principal.ID,
		principal.ProviderName,
		principal.ProviderUserID,
		accessTokenRespBytes,
		userProfileBytes,
		createdAt,
		updatedAt,
	)

	_, err = p.sqlExecutor.ExecWith(builder)
	if err != nil {
		if db.IsUniqueViolated(err) {
			err = skydb.ErrUserDuplicated
		}
	}

	return
}

func (p *providerImpl) UpdatePrincipal(principal *Principal) (err error) {
	var (
		updatedAt *time.Time
	)
	updatedAt = principal.UpdatedAt
	if updatedAt != nil && updatedAt.IsZero() {
		updatedAt = nil
	}

	var accessTokenRespBytes []byte
	accessTokenRespBytes, err = json.Marshal(principal.AccessTokenResp)
	if err != nil {
		return
	}

	var userProfileBytes []byte
	userProfileBytes, err = json.Marshal(principal.UserProfile)
	if err != nil {
		return
	}

	builder := p.sqlBuilder.Update(p.sqlBuilder.FullTableName("provider_oauth")).
		Set("token_response", accessTokenRespBytes).
		Set("profile", userProfileBytes).
		Set("_updated_at", updatedAt).
		Where("oauth_provider = ? and principal_id = ?", principal.ProviderName, principal.ID)

	result, err := p.sqlExecutor.ExecWith(builder)
	if err != nil {
		if db.IsUniqueViolated(err) {
			err = skydb.ErrUserDuplicated
		}
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return skydb.ErrUserNotFound
	} else if rowsAffected > 1 {
		panic(fmt.Errorf("want 1 rows updated, got %v", rowsAffected))
	}

	return nil
}

func (p *providerImpl) DeletePrincipal(principalProviderName string, principal *Principal) (err error) {
	// Delete provider_oauth
	builder := p.sqlBuilder.Delete(p.sqlBuilder.FullTableName("provider_oauth")).
		Where("oauth_provider = ? and principal_id = ?", principalProviderName, principal.ID)

	result, err := p.sqlExecutor.ExecWith(builder)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return skydb.ErrUserNotFound
	} else if rowsAffected > 1 {
		panic(fmt.Errorf("want 1 rows deleted, got %v", rowsAffected))
	}

	// Delete principal
	builder = p.sqlBuilder.Delete(p.sqlBuilder.FullTableName("principal")).
		Where("id = ? and provider = ?", principal.ID, providerName)

	result, err = p.sqlExecutor.ExecWith(builder)
	if err != nil {
		return err
	}

	rowsAffected, err = result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return skydb.ErrUserNotFound
	} else if rowsAffected > 1 {
		panic(fmt.Errorf("want 1 rows deleted, got %v", rowsAffected))
	}

	return
}

func (p *providerImpl) GetPrincipalsByUserID(userID string) (principals []*Principal, err error) {
	builder := p.sqlBuilder.Select("p.id", "oauth.oauth_provider", "oauth.provider_user_id", "oauth.profile").
		From(fmt.Sprintf("%s as p", p.sqlBuilder.FullTableName("principal"))).
		Join(p.sqlBuilder.FullTableName("provider_oauth")+" AS oauth ON p.id = oauth.principal_id").
		Where("p.user_id = ? AND p.provider = 'oauth'", userID)
	rows, err := p.sqlExecutor.QueryWith(builder)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var profileDataBytes []byte
		var principal Principal
		principal.UserID = userID
		if err = rows.Scan(
			&principal.ID,
			&principal.ProviderName,
			&principal.ProviderUserID,
			&profileDataBytes,
		); err != nil {
			return
		}

		err = json.Unmarshal(profileDataBytes, &principal.UserProfile)
		if err != nil {
			return
		}

		principals = append(principals, &principal)
	}

	return
}

// this ensures that our structure conform to certain interfaces.
var (
	_ Provider = &providerImpl{}
)

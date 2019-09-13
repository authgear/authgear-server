package oauth

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/lib/pq"
	"github.com/sirupsen/logrus"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/skydb"
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

func (p *providerImpl) scan(scanner db.Scanner, principal *Principal) error {
	var tokenBytes []byte
	var profileBytes []byte
	var providerKeysBytes []byte
	var claimsValueBytes []byte

	err := scanner.Scan(
		&principal.ID,
		&principal.UserID,
		&principal.ProviderType,
		&providerKeysBytes,
		&principal.ProviderUserID,
		&tokenBytes,
		&profileBytes,
		&claimsValueBytes,
		&principal.CreatedAt,
		&principal.UpdatedAt,
	)
	if err != nil {
		return err
	}

	err = json.Unmarshal(tokenBytes, &principal.AccessTokenResp)
	if err != nil {
		return err
	}

	err = json.Unmarshal(profileBytes, &principal.UserProfile)
	if err != nil {
		return err
	}

	err = json.Unmarshal(providerKeysBytes, &principal.ProviderKeys)
	if err != nil {
		return err
	}

	err = json.Unmarshal(claimsValueBytes, &principal.ClaimsValue)
	if err != nil {
		return err
	}

	return nil
}

func (p *providerImpl) GetPrincipalByProvider(options GetByProviderOptions) (*Principal, error) {
	if options.ProviderKeys == nil {
		options.ProviderKeys = map[string]interface{}{}
	}

	principal := Principal{}
	providerKeysBytes, err := json.Marshal(options.ProviderKeys)
	if err != nil {
		return nil, err
	}

	builder := p.sqlBuilder.Tenant().
		Select(
			"p.id",
			"p.user_id",
			"o.provider_type",
			"o.provider_keys",
			"o.provider_user_id",
			"o.token_response",
			"o.profile",
			"o.claims",
			"o._created_at",
			"o._updated_at",
		).
		From(p.sqlBuilder.FullTableName("principal"), "p").
		Join(p.sqlBuilder.FullTableName("provider_oauth"), "o", "p.id = o.principal_id").
		Where(
			"o.provider_type = ? AND o.provider_keys = ? AND o.provider_user_id = ?",
			options.ProviderType, providerKeysBytes, options.ProviderUserID)

	row := p.sqlExecutor.QueryRowWith(builder)

	err = p.scan(row, &principal)

	if err == sql.ErrNoRows {
		err = skydb.ErrUserNotFound
	}

	if err != nil {
		return nil, err
	}

	return &principal, nil
}

func (p *providerImpl) GetPrincipalByUser(options GetByUserOptions) (*Principal, error) {
	if options.ProviderKeys == nil {
		options.ProviderKeys = map[string]interface{}{}
	}

	principal := Principal{}
	providerKeysBytes, err := json.Marshal(options.ProviderKeys)
	if err != nil {
		return nil, err
	}

	builder := p.sqlBuilder.Tenant().
		Select(
			"p.id",
			"p.user_id",
			"o.provider_type",
			"o.provider_keys",
			"o.provider_user_id",
			"o.token_response",
			"o.profile",
			"o.claims",
			"o._created_at",
			"o._updated_at",
		).
		From(p.sqlBuilder.FullTableName("principal"), "p").
		Join(p.sqlBuilder.FullTableName("provider_oauth"), "o", "p.id = o.principal_id").
		Where(
			"o.provider_type = ? AND o.provider_keys = ? AND p.user_id = ?",
			options.ProviderType, providerKeysBytes, options.UserID)

	row := p.sqlExecutor.QueryRowWith(builder)

	err = p.scan(row, &principal)

	if err == sql.ErrNoRows {
		err = skydb.ErrUserNotFound
	}

	if err != nil {
		return nil, err
	}

	return &principal, nil
}

func (p *providerImpl) CreatePrincipal(principal *Principal) (err error) {
	// Create principal
	builder := p.sqlBuilder.Tenant().
		Insert(p.sqlBuilder.FullTableName("principal")).
		Columns(
			"id",
			"provider",
			"user_id",
		).
		Values(
			principal.ID,
			coreAuth.PrincipalTypeOAuth,
			principal.UserID,
		)

	_, err = p.sqlExecutor.ExecWith(builder)
	if err != nil {
		return
	}

	accessTokenRespBytes, err := json.Marshal(principal.AccessTokenResp)
	if err != nil {
		return
	}
	userProfileBytes, err := json.Marshal(principal.UserProfile)
	if err != nil {
		return
	}
	claimsValueBytes, err := json.Marshal(principal.ClaimsValue)
	if err != nil {
		return
	}
	providerKeysBytes, err := json.Marshal(principal.ProviderKeys)
	if err != nil {
		return
	}

	builder = p.sqlBuilder.Tenant().
		Insert(p.sqlBuilder.FullTableName("provider_oauth")).
		Columns(
			"principal_id",
			"provider_type",
			"provider_keys",
			"provider_user_id",
			"token_response",
			"profile",
			"claims",
			"_created_at",
			"_updated_at",
		).
		Values(
			principal.ID,
			principal.ProviderType,
			providerKeysBytes,
			principal.ProviderUserID,
			accessTokenRespBytes,
			userProfileBytes,
			claimsValueBytes,
			principal.CreatedAt,
			principal.UpdatedAt,
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
	accessTokenRespBytes, err := json.Marshal(principal.AccessTokenResp)
	if err != nil {
		return
	}

	userProfileBytes, err := json.Marshal(principal.UserProfile)
	if err != nil {
		return
	}

	claimsValueBytes, err := json.Marshal(principal.ClaimsValue)
	if err != nil {
		return
	}

	builder := p.sqlBuilder.Tenant().
		Update(p.sqlBuilder.FullTableName("provider_oauth")).
		Set("token_response", accessTokenRespBytes).
		Set("profile", userProfileBytes).
		Set("claims", claimsValueBytes).
		Set("_updated_at", principal.UpdatedAt).
		Where("principal_id = ?", principal.ID)

	result, err := p.sqlExecutor.ExecWith(builder)
	if err != nil {
		if db.IsUniqueViolated(err) {
			err = skydb.ErrUserDuplicated
			return
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

func (p *providerImpl) DeletePrincipal(principal *Principal) (err error) {
	// Delete provider_oauth
	builder := p.sqlBuilder.Tenant().
		Delete(p.sqlBuilder.FullTableName("provider_oauth")).
		Where("principal_id = ?", principal.ID)

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
	builder = p.sqlBuilder.Tenant().
		Delete(p.sqlBuilder.FullTableName("principal")).
		Where("id = ?", principal.ID)

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
	builder := p.sqlBuilder.Tenant().
		Select(
			"p.id",
			"p.user_id",
			"o.provider_type",
			"o.provider_keys",
			"o.provider_user_id",
			"o.token_response",
			"o.profile",
			"o.claims",
			"o._created_at",
			"o._updated_at",
		).
		From(p.sqlBuilder.FullTableName("principal"), "p").
		Join(p.sqlBuilder.FullTableName("provider_oauth"), "o", "p.id = o.principal_id").
		Where(
			"p.user_id = ? AND p.provider = ?",
			userID,
			coreAuth.PrincipalTypeOAuth)

	rows, err := p.sqlExecutor.QueryWith(builder)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var principal Principal
		err = p.scan(rows, &principal)
		if err != nil {
			return
		}
		principals = append(principals, &principal)
	}

	if len(principals) == 0 {
		err = skydb.ErrUserNotFound
		return
	}

	return
}

func (p *providerImpl) GetPrincipalsByClaim(claimName string, claimValue string) (principals []*Principal, err error) {
	builder := p.sqlBuilder.Tenant().
		Select(
			"p.id",
			"p.user_id",
			"o.provider_type",
			"o.provider_keys",
			"o.provider_user_id",
			"o.token_response",
			"o.profile",
			"o.claims",
			"o._created_at",
			"o._updated_at",
		).
		From(p.sqlBuilder.FullTableName("principal"), "p").
		Join(p.sqlBuilder.FullTableName("provider_oauth"), "o", "p.id = o.principal_id").
		Where("(o.claims #>> ?) = ?", pq.Array([]string{claimName}), claimValue)

	rows, err := p.sqlExecutor.QueryWith(builder)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var principal Principal
		err = p.scan(rows, &principal)
		if err != nil {
			return
		}
		principals = append(principals, &principal)
	}

	if len(principals) == 0 {
		err = skydb.ErrUserNotFound
		return
	}

	return
}

func (p *providerImpl) ID() string {
	return string(coreAuth.PrincipalTypeOAuth)
}

func (p *providerImpl) GetPrincipalByID(principalID string) (principal.Principal, error) {
	builder := p.sqlBuilder.Tenant().
		Select(
			"p.id",
			"p.user_id",
			"o.provider_type",
			"o.provider_keys",
			"o.provider_user_id",
			"o.token_response",
			"o.profile",
			"o.claims",
			"o._created_at",
			"o._updated_at",
		).
		From(p.sqlBuilder.FullTableName("principal"), "p").
		Join(p.sqlBuilder.FullTableName("provider_oauth"), "o", "p.id = o.principal_id").
		Where("p.id = ?", principalID)

	scanner := p.sqlExecutor.QueryRowWith(builder)

	var principal Principal
	err := p.scan(scanner, &principal)

	if err == sql.ErrNoRows {
		err = skydb.ErrUserNotFound
	}

	if err != nil {
		return nil, err
	}

	return &principal, nil
}

func (p *providerImpl) ListPrincipalsByUserID(userID string) ([]principal.Principal, error) {
	principals, err := p.GetPrincipalsByUserID(userID)
	if err != nil {
		if err == skydb.ErrUserNotFound {
			return nil, nil
		}
		return nil, err
	}

	genericPrincipals := []principal.Principal{}
	for _, principal := range principals {
		genericPrincipals = append(genericPrincipals, principal)
	}

	return genericPrincipals, nil
}

func (p *providerImpl) ListPrincipalsByClaim(claimName string, claimValue string) ([]principal.Principal, error) {
	principals, err := p.GetPrincipalsByClaim(claimName, claimValue)
	if err != nil {
		if err == skydb.ErrUserNotFound {
			return nil, nil
		}
		return nil, err
	}

	genericPrincipals := []principal.Principal{}
	for _, principal := range principals {
		genericPrincipals = append(genericPrincipals, principal)
	}

	return genericPrincipals, nil
}

// this ensures that our structure conform to certain interfaces.
var (
	_ Provider = &providerImpl{}
)

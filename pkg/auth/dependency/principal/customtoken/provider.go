package customtoken

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/dgrijalva/jwt-go"
	"github.com/lib/pq"

	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/skydb"
)

type providerImpl struct {
	sqlBuilder        db.SQLBuilder
	sqlExecutor       db.SQLExecutor
	logger            *logrus.Entry
	customTokenConfig config.CustomTokenConfiguration
}

func newProvider(
	builder db.SQLBuilder,
	executor db.SQLExecutor,
	logger *logrus.Entry,
	customTokenConfig config.CustomTokenConfiguration,
) *providerImpl {
	return &providerImpl{
		sqlBuilder:        builder,
		sqlExecutor:       executor,
		logger:            logger,
		customTokenConfig: customTokenConfig,
	}
}

func NewProvider(
	builder db.SQLBuilder,
	executor db.SQLExecutor,
	logger *logrus.Entry,
	customTokenConfig config.CustomTokenConfiguration,
) Provider {
	return newProvider(builder, executor, logger, customTokenConfig)
}

func (p *providerImpl) scan(scanner db.Scanner, principal *Principal) error {
	var rawProfileBytes []byte
	var claimsValueBytes []byte
	err := scanner.Scan(
		&principal.ID,
		&principal.UserID,
		&principal.TokenPrincipalID,
		&rawProfileBytes,
		&claimsValueBytes,
	)
	if err != nil {
		return err
	}

	err = json.Unmarshal(rawProfileBytes, &principal.RawProfile)
	if err != nil {
		return err
	}

	err = json.Unmarshal(claimsValueBytes, &principal.ClaimsValue)
	if err != nil {
		return err
	}

	return nil
}

func (p *providerImpl) Decode(tokenString string) (claims SSOCustomTokenClaims, err error) {
	_, err = jwt.ParseWithClaims(
		tokenString,
		&claims,
		func(token *jwt.Token) (interface{}, error) {
			err := errors.New("invalid token: invalid signature method")
			method, ok := token.Method.(*jwt.SigningMethodHMAC)
			if !ok {
				return nil, err
			}
			if method != jwt.SigningMethodHS256 {
				return nil, err
			}
			return []byte(p.customTokenConfig.Secret), nil
		},
	)

	return
}

func (p *providerImpl) CreatePrincipal(principal *Principal) (err error) {
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

	rawProfileBytes, err := json.Marshal(principal.RawProfile)
	if err != nil {
		return
	}

	claimsBytes, err := json.Marshal(principal.ClaimsValue)
	if err != nil {
		return
	}

	builder = p.sqlBuilder.Insert(p.sqlBuilder.FullTableName("provider_custom_token")).Columns(
		"principal_id",
		"raw_profile",
		"token_principal_id",
		"claims",
	).Values(
		principal.ID,
		rawProfileBytes,
		principal.TokenPrincipalID,
		claimsBytes,
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
	rawProfileBytes, err := json.Marshal(principal.RawProfile)
	if err != nil {
		return
	}

	claimsBytes, err := json.Marshal(principal.ClaimsValue)
	if err != nil {
		return
	}

	builder := p.sqlBuilder.Update(p.sqlBuilder.FullTableName("provider_custom_token")).
		Set("raw_profile", rawProfileBytes).
		Set("claims", claimsBytes).
		Where("principal_id = ?", principal.ID)

	result, err := p.sqlExecutor.ExecWith(builder)
	if err != nil {
		return
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

func (p *providerImpl) GetPrincipalByTokenPrincipalID(tokenPrincipalID string) (*Principal, error) {
	principal := Principal{}
	builder := p.sqlBuilder.Select(
		"p.id",
		"p.user_id",
		"ct.token_principal_id",
		"ct.raw_profile",
		"ct.claims",
	).
		From(fmt.Sprintf("%s AS p", p.sqlBuilder.FullTableName("principal"))).
		Join(fmt.Sprintf("%s AS ct ON p.id = ct.principal_id", p.sqlBuilder.FullTableName("provider_custom_token"))).
		Where("ct.token_principal_id = ?", tokenPrincipalID)

	scanner := p.sqlExecutor.QueryRowWith(builder)

	err := p.scan(scanner, &principal)
	if err == sql.ErrNoRows {
		err = skydb.ErrUserNotFound
	}

	if err != nil {
		return nil, err
	}

	return &principal, nil
}

func (p *providerImpl) ID() string {
	return providerName
}

func (p *providerImpl) GetPrincipalByID(principalID string) (principal.Principal, error) {
	principal := Principal{}
	builder := p.sqlBuilder.Select(
		"p.id",
		"p.user_id",
		"ct.token_principal_id",
		"ct.raw_profile",
		"ct.claims",
	).
		From(fmt.Sprintf("%s AS p", p.sqlBuilder.FullTableName("principal"))).
		Join(fmt.Sprintf("%s AS ct ON p.id = ct.principal_id", p.sqlBuilder.FullTableName("provider_custom_token"))).
		Where("p.id = ?", principalID)

	scanner := p.sqlExecutor.QueryRowWith(builder)

	err := p.scan(scanner, &principal)
	if err == sql.ErrNoRows {
		err = skydb.ErrUserNotFound
	}

	if err != nil {
		return nil, err
	}

	return &principal, nil
}

func (p *providerImpl) ListPrincipalsByUserID(userID string) (principals []principal.Principal, err error) {
	builder := p.sqlBuilder.Select(
		"p.id",
		"p.user_id",
		"ct.token_principal_id",
		"ct.raw_profile",
		"ct.claims",
	).
		From(fmt.Sprintf("%s AS p", p.sqlBuilder.FullTableName("principal"))).
		Join(fmt.Sprintf("%s AS ct ON p.id = ct.principal_id", p.sqlBuilder.FullTableName("provider_custom_token"))).
		Where("p.user_id = ?", userID)

	rows, err := p.sqlExecutor.QueryWith(builder)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		principal := Principal{}
		err = p.scan(rows, &principal)
		if err != nil {
			return
		}
		principals = append(principals, &principal)
	}

	return
}

func (p *providerImpl) ListPrincipalsByClaim(claimName string, claimValue string) (principals []principal.Principal, err error) {
	builder := p.sqlBuilder.Select(
		"p.id",
		"p.user_id",
		"ct.token_principal_id",
		"ct.raw_profile",
		"ct.claims",
	).
		From(fmt.Sprintf("%s AS p", p.sqlBuilder.FullTableName("principal"))).
		Join(fmt.Sprintf("%s AS ct ON p.id = ct.principal_id", p.sqlBuilder.FullTableName("provider_custom_token"))).
		Where("(ct.claims #>> ?) = ?", pq.Array([]string{claimName}), claimValue)

	rows, err := p.sqlExecutor.QueryWith(builder)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		principal := Principal{}
		err = p.scan(rows, &principal)
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

package customtoken

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/dgrijalva/jwt-go"

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

func (p providerImpl) Decode(tokenString string) (claims SSOCustomTokenClaims, err error) {
	_, err = jwt.ParseWithClaims(
		tokenString,
		&claims,
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("fails to parse token")
			}
			return []byte(p.customTokenConfig.Secret), nil
		},
	)

	return
}

func (p providerImpl) CreatePrincipal(principal Principal) (err error) {
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

	builder = p.sqlBuilder.Insert(p.sqlBuilder.FullTableName("provider_custom_token")).Columns(
		"principal_id",
		"token_principal_id",
	).Values(
		principal.ID,
		principal.TokenPrincipalID,
	)

	_, err = p.sqlExecutor.ExecWith(builder)
	if err != nil {
		if db.IsUniqueViolated(err) {
			err = skydb.ErrUserDuplicated
		}
	}

	return
}

func (p providerImpl) GetPrincipalByTokenPrincipalID(tokenPrincipalID string) (*Principal, error) {
	principal := Principal{}
	principal.TokenPrincipalID = tokenPrincipalID

	builder := p.sqlBuilder.Select("p.id", "p.user_id").
		From(fmt.Sprintf("%s as p", p.sqlBuilder.FullTableName("principal"))).
		Join(p.sqlBuilder.FullTableName("provider_custom_token")+" AS ct ON p.id = ct.principal_id").
		Where("ct.token_principal_id = ? AND p.provider = 'custom_token'", tokenPrincipalID)
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

func (p providerImpl) ID() string {
	return providerName
}

func (p providerImpl) GetPrincipalByID(principalID string) (principal.Principal, error) {
	principal := Principal{ID: principalID}

	builder := p.sqlBuilder.Select("ct.token_principal_id", "p.user_id").
		From(fmt.Sprintf("%s as p", p.sqlBuilder.FullTableName("principal"))).
		Join(p.sqlBuilder.FullTableName("provider_custom_token")+" AS ct ON p.id = ct.principal_id").
		Where("p.id = ? AND p.provider = 'custom_token'", principalID)
	scanner := p.sqlExecutor.QueryRowWith(builder)

	err := scanner.Scan(
		&principal.TokenPrincipalID,
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

func (p providerImpl) ListPrincipalsByUserID(userID string) (principals []principal.Principal, err error) {
	builder := p.sqlBuilder.Select("p.id", "ct.token_principal_id").
		From(fmt.Sprintf("%s as p", p.sqlBuilder.FullTableName("principal"))).
		Join(p.sqlBuilder.FullTableName("provider_custom_token")+" AS ct ON p.id = ct.principal_id").
		Where("p.user_id = ? AND p.provider = 'custom_token'", userID)
	rows, err := p.sqlExecutor.QueryWith(builder)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		principal := Principal{UserID: userID}
		if err = rows.Scan(
			&principal.ID,
			&principal.TokenPrincipalID,
		); err != nil {
			return
		}

		principals = append(principals, &principal)
	}

	return
}

func (p providerImpl) DeriveClaims(_ principal.Principal) principal.Claims {
	// TODO(sso): return custom token email
	return principal.Claims{}
}

// this ensures that our structure conform to certain interfaces.
var (
	_ Provider = &providerImpl{}
)

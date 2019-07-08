package customtoken

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"

	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/skydb"
)

type providerImpl struct {
	sqlBuilder  db.SQLBuilder
	sqlExecutor db.SQLExecutor
	logger      *logrus.Entry
	secret      string
}

func newProvider(
	builder db.SQLBuilder,
	executor db.SQLExecutor,
	logger *logrus.Entry,
	secret string,
) *providerImpl {
	return &providerImpl{
		sqlBuilder:  builder,
		sqlExecutor: executor,
		logger:      logger,
		secret:      secret,
	}
}

func NewProvider(
	builder db.SQLBuilder,
	executor db.SQLExecutor,
	logger *logrus.Entry,
	secret string,
) Provider {
	return newProvider(builder, executor, logger, secret)
}

func (p providerImpl) Decode(tokenString string) (claims SSOCustomTokenClaims, err error) {
	_, err = jwt.ParseWithClaims(
		tokenString,
		&claims,
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("fails to parse token")
			}
			return []byte(p.secret), nil
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

// this ensures that our structure conform to certain interfaces.
var (
	_ Provider = &providerImpl{}
)

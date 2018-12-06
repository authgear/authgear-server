package oauth

import (
	"database/sql"
	"fmt"

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

func (p providerImpl) GetPrincipalByUserID(providerName string, userID string) (*Principal, error) {
	principal := Principal{}
	principal.ProviderName = providerName
	principal.ProviderUserID = userID

	builder := p.sqlBuilder.Select("p.id", "p.user_id").
		From(fmt.Sprintf("%s as p", p.sqlBuilder.FullTableName("principal"))).
		Join(p.sqlBuilder.FullTableName("provider_oauth")+" AS oauth ON p.id = oauth.principal_id").
		Where("oauth.oauth_provider = ? AND oauth.user_id = ? AND p.provider = 'oauth'", providerName, userID)
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

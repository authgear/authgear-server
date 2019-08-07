package anonymous

import (
	"database/sql"

	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/skydb"
)

type providerImpl struct {
	sqlBuilder  db.SQLBuilder
	sqlExecutor db.SQLExecutor
	logger      *logrus.Entry
}

func newProvider(builder db.SQLBuilder, executor db.SQLExecutor, logger *logrus.Entry) *providerImpl {
	return &providerImpl{
		sqlBuilder:  builder,
		sqlExecutor: executor,
		logger:      logger,
	}
}

func NewProvider(builder db.SQLBuilder, executor db.SQLExecutor, logger *logrus.Entry) Provider {
	return newProvider(builder, executor, logger)
}

func (p providerImpl) CreatePrincipal(principal Principal) (err error) {
	// TODO: log

	// Create principal
	builder := p.sqlBuilder.Insert(p.sqlBuilder.FullTableName("principal")).Columns(
		"id",
		"provider",
		"user_id",
	).Values(
		principal.ID,
		providerAnonymous,
		principal.UserID,
	)

	_, err = p.sqlExecutor.ExecWith(builder)
	if err != nil {
		return
	}

	return
}

func (p providerImpl) ID() string {
	return providerAnonymous
}

func (p providerImpl) GetPrincipalByID(principalID string) (principal.Principal, error) {
	principal := Principal{ID: principalID}

	builder := p.sqlBuilder.Select("user_id").
		From(p.sqlBuilder.FullTableName("principal")).
		Where("id = ? AND provider = ?", principalID, providerAnonymous)
	scanner := p.sqlExecutor.QueryRowWith(builder)

	err := scanner.Scan(&principal.UserID)

	if err == sql.ErrNoRows {
		err = skydb.ErrUserNotFound
	}

	if err != nil {
		return nil, err
	}

	return &principal, nil
}

func (p providerImpl) ListPrincipalsByUserID(userID string) (principals []principal.Principal, err error) {
	builder := p.sqlBuilder.Select("id").
		From(p.sqlBuilder.FullTableName("principal")).
		Where("user_id = ? AND p.provider = ?", userID, providerAnonymous)
	rows, err := p.sqlExecutor.QueryWith(builder)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		principal := Principal{UserID: userID}
		if err = rows.Scan(&principal.ID); err != nil {
			return
		}

		principals = append(principals, &principal)
	}

	return
}

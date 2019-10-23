package anonymous

import (
	"database/sql"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/errors"
)

type providerImpl struct {
	sqlBuilder  db.SQLBuilder
	sqlExecutor db.SQLExecutor
}

func newProvider(builder db.SQLBuilder, executor db.SQLExecutor) *providerImpl {
	return &providerImpl{
		sqlBuilder:  builder,
		sqlExecutor: executor,
	}
}

func NewProvider(builder db.SQLBuilder, executor db.SQLExecutor) Provider {
	return newProvider(builder, executor)
}

func (p providerImpl) CreatePrincipal(principal Principal) (err error) {
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
			providerAnonymous,
			principal.UserID,
		)

	_, err = p.sqlExecutor.ExecWith(builder)
	if err != nil {
		return errors.HandledWithMessage(err, "fail to create principal")
	}

	return
}

func (p providerImpl) ID() string {
	return providerAnonymous
}

func (p providerImpl) GetPrincipalByID(principalID string) (principal.Principal, error) {
	builder := p.sqlBuilder.Tenant().
		Select("user_id").
		From(p.sqlBuilder.FullTableName("principal")).
		Where("id = ? AND provider = ?", principalID, providerAnonymous)
	scanner, err := p.sqlExecutor.QueryRowWith(builder)
	if err != nil {
		return nil, errors.HandledWithMessage(err, "fail to get principal by ID")
	}

	pp := Principal{ID: principalID}
	err = scanner.Scan(&pp.UserID)

	if err == sql.ErrNoRows {
		return nil, principal.ErrNotFound
	} else if err != nil {
		return nil, errors.HandledWithMessage(err, "fail to get principal by ID")
	}

	return &pp, nil
}

func (p providerImpl) ListPrincipalsByUserID(userID string) (principals []principal.Principal, err error) {
	builder := p.sqlBuilder.Tenant().
		Select("id").
		From(p.sqlBuilder.FullTableName("principal")).
		Where("user_id = ? AND p.provider = ?", userID, providerAnonymous)
	rows, err := p.sqlExecutor.QueryWith(builder)
	if err != nil {
		return nil, errors.HandledWithMessage(err, "fail to get principal by user ID")
	}
	defer rows.Close()

	for rows.Next() {
		principal := Principal{UserID: userID}
		if err = rows.Scan(&principal.ID); err != nil {
			return nil, errors.HandledWithMessage(err, "fail to get principal by user ID")
		}

		principals = append(principals, &principal)
	}

	return
}

func (p providerImpl) ListPrincipalsByClaim(claimName string, claimValue string) (principals []principal.Principal, err error) {
	return
}

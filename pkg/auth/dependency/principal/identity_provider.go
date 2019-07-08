package principal

import (
	"database/sql"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/skydb"
)

type IdentityProvider struct {
	sqlBuilder  db.SQLBuilder
	sqlExecutor db.SQLExecutor
	providers   []Provider
}

func NewIdentityProvider(builder db.SQLBuilder, executor db.SQLExecutor, providers ...Provider) *IdentityProvider {
	return &IdentityProvider{builder, executor, providers}
}

func (p *IdentityProvider) ListPrincipalsByUserID(userID string) ([]Principal, error) {
	principals := []Principal{}
	for _, provider := range p.providers {
		providerPrincipals, err := provider.ListPrincipalsByUserID(userID)
		if err != nil {
			return nil, err
		}
		principals = append(principals, providerPrincipals...)
	}
	return principals, nil
}

func (p *IdentityProvider) GetPrincipalByID(principalID string) (Principal, error) {
	var providerID string

	builder := p.sqlBuilder.Select("provider").
		From(p.sqlBuilder.FullTableName("principal")).
		Where("id = ?", principalID)
	scanner := p.sqlExecutor.QueryRowWith(builder)

	err := scanner.Scan(&providerID)

	if err == sql.ErrNoRows {
		err = skydb.ErrUserNotFound
	}

	if err != nil {
		return nil, err
	}

	for _, provider := range p.providers {
		if provider.ID() == providerID {
			return provider.GetPrincipalByID(principalID)
		}
	}

	return nil, skydb.ErrUserNotFound
}

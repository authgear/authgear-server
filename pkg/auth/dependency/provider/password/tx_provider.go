package password

import (
	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/core/db"
)

type safeProviderImpl struct {
	impl      *providerImpl
	txContext db.SafeTxContext
}

func NewSafeProvider(
	builder db.SQLBuilder,
	executor db.SQLExecutor,
	logger *logrus.Entry,
	authRecordKeys [][]string,
	txContext db.SafeTxContext,
) Provider {
	return &safeProviderImpl{
		impl:      newProvider(builder, executor, logger, authRecordKeys),
		txContext: txContext,
	}
}

func (p *safeProviderImpl) IsAuthDataValid(authData map[string]interface{}) bool {
	p.txContext.EnsureTx()
	return p.impl.IsAuthDataValid(authData)
}

func (p *safeProviderImpl) CreatePrincipalsByAuthData(authInfoID string, password string, authData map[string]interface{}) error {
	p.txContext.EnsureTx()
	return p.impl.CreatePrincipalsByAuthData(authInfoID, password, authData)
}

func (p *safeProviderImpl) CreatePrincipal(principal Principal) error {
	p.txContext.EnsureTx()
	return p.impl.CreatePrincipal(principal)
}

func (p *safeProviderImpl) GetPrincipalByAuthData(authData map[string]interface{}, principal *Principal) error {
	p.txContext.EnsureTx()
	return p.impl.GetPrincipalByAuthData(authData, principal)
}

func (p *safeProviderImpl) GetPrincipalByUserID(userID string) ([]*Principal, error) {
	p.txContext.EnsureTx()
	return p.impl.GetPrincipalByUserID(userID)
}

func (p *safeProviderImpl) UpdatePrincipal(principal Principal) error {
	p.txContext.EnsureTx()
	return p.impl.UpdatePrincipal(principal)
}

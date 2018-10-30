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
	txContext db.SafeTxContext,
) Provider {
	return &safeProviderImpl{
		impl:      newProvider(builder, executor, logger),
		txContext: txContext,
	}
}

func (p *safeProviderImpl) CreatePrincipal(principal Principal) error {
	p.txContext.EnsureTx()
	return p.CreatePrincipal(principal)
}

func (p *safeProviderImpl) GetPrincipalByAuthData(authData map[string]interface{}, principal *Principal) error {
	p.txContext.EnsureTx()
	return p.GetPrincipalByAuthData(authData, principal)
}

func (p *safeProviderImpl) GetPrincipalByUserID(userID string, principal *Principal) error {
	p.txContext.EnsureTx()
	return p.GetPrincipalByUserID(userID, principal)
}

func (p *safeProviderImpl) UpdatePrincipal(principal Principal) error {
	p.txContext.EnsureTx()
	return p.UpdatePrincipal(principal)
}

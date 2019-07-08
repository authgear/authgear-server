package customtoken

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
	secret string,
	txContext db.SafeTxContext,
) Provider {
	return &safeProviderImpl{
		impl:      newProvider(builder, executor, logger, secret),
		txContext: txContext,
	}
}

func (p *safeProviderImpl) Decode(tokenString string) (SSOCustomTokenClaims, error) {
	return p.impl.Decode(tokenString)
}

func (p *safeProviderImpl) CreatePrincipal(principal Principal) error {
	p.txContext.EnsureTx()
	return p.impl.CreatePrincipal(principal)
}

func (p *safeProviderImpl) GetPrincipalByTokenPrincipalID(tokenPrincipalID string) (*Principal, error) {
	p.txContext.EnsureTx()
	return p.impl.GetPrincipalByTokenPrincipalID(tokenPrincipalID)
}

package oauth

import (
	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
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

func (p *safeProviderImpl) GetPrincipalByProvider(options GetByProviderOptions) (*Principal, error) {
	p.txContext.EnsureTx()
	return p.impl.GetPrincipalByProvider(options)
}

func (p *safeProviderImpl) GetPrincipalByUser(options GetByUserOptions) (*Principal, error) {
	p.txContext.EnsureTx()
	return p.impl.GetPrincipalByUser(options)
}

func (p *safeProviderImpl) CreatePrincipal(principal *Principal) error {
	p.txContext.EnsureTx()
	return p.impl.CreatePrincipal(principal)
}

func (p *safeProviderImpl) UpdatePrincipal(principal *Principal) error {
	p.txContext.EnsureTx()
	return p.impl.UpdatePrincipal(principal)
}

func (p *safeProviderImpl) DeletePrincipal(principal *Principal) error {
	p.txContext.EnsureTx()
	return p.impl.DeletePrincipal(principal)
}

func (p *safeProviderImpl) GetPrincipalsByUserID(userID string) ([]*Principal, error) {
	p.txContext.EnsureTx()
	return p.impl.GetPrincipalsByUserID(userID)
}

func (p *safeProviderImpl) ID() string {
	p.txContext.EnsureTx()
	return p.impl.ID()
}

func (p *safeProviderImpl) GetPrincipalByID(principalID string) (principal.Principal, error) {
	p.txContext.EnsureTx()
	return p.impl.GetPrincipalByID(principalID)
}

func (p *safeProviderImpl) ListPrincipalsByUserID(userID string) ([]principal.Principal, error) {
	p.txContext.EnsureTx()
	return p.impl.ListPrincipalsByUserID(userID)
}

func (p *safeProviderImpl) DeriveClaims(principal principal.Principal) principal.Claims {
	p.txContext.EnsureTx()
	return p.impl.DeriveClaims(principal)
}

var (
	_ Provider = &safeProviderImpl{}
)

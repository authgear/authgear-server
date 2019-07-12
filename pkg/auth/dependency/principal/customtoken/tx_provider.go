package customtoken

import (
	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"

	"github.com/skygeario/skygear-server/pkg/core/config"
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
	customTokenConfig config.CustomTokenConfiguration,
	txContext db.SafeTxContext,
) Provider {
	return &safeProviderImpl{
		impl:      newProvider(builder, executor, logger, customTokenConfig),
		txContext: txContext,
	}
}

func (p *safeProviderImpl) Decode(tokenString string) (SSOCustomTokenClaims, error) {
	return p.impl.Decode(tokenString)
}

func (p *safeProviderImpl) CreatePrincipal(principal *Principal) error {
	p.txContext.EnsureTx()
	return p.impl.CreatePrincipal(principal)
}

func (p *safeProviderImpl) UpdatePrincipal(principal *Principal) error {
	p.txContext.EnsureTx()
	return p.impl.UpdatePrincipal(principal)
}

func (p *safeProviderImpl) GetPrincipalByTokenPrincipalID(tokenPrincipalID string) (*Principal, error) {
	p.txContext.EnsureTx()
	return p.impl.GetPrincipalByTokenPrincipalID(tokenPrincipalID)
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

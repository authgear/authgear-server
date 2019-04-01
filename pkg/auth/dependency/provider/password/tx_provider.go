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
	loginIDMetadataKeys [][]string,
	passwordHistoryEnabled bool,
	txContext db.SafeTxContext,
) Provider {
	return &safeProviderImpl{
		impl:      newProvider(builder, executor, logger, loginIDMetadataKeys, passwordHistoryEnabled),
		txContext: txContext,
	}
}

func (p *safeProviderImpl) IsAuthDataValid(authData map[string]string) bool {
	p.txContext.EnsureTx()
	return p.impl.IsAuthDataValid(authData)
}

func (p *safeProviderImpl) IsAuthDataMatching(authData map[string]string) bool {
	p.txContext.EnsureTx()
	return p.impl.IsAuthDataMatching(authData)
}

func (p *safeProviderImpl) GetLoginIDMetadataFlattenedKeys() []string {
	p.txContext.EnsureTx()
	return p.impl.GetLoginIDMetadataFlattenedKeys()
}

func (p *safeProviderImpl) CreatePrincipalsByAuthData(authInfoID string, password string, authData map[string]string) error {
	p.txContext.EnsureTx()
	return p.impl.CreatePrincipalsByAuthData(authInfoID, password, authData)
}

func (p *safeProviderImpl) CreatePrincipal(principal Principal) error {
	p.txContext.EnsureTx()
	return p.impl.CreatePrincipal(principal)
}

func (p *safeProviderImpl) GetPrincipalByAuthData(authData map[string]string, principal *Principal) (err error) {
	p.txContext.EnsureTx()
	return p.impl.GetPrincipalByAuthData(authData, principal)
}

func (p *safeProviderImpl) GetPrincipalsByUserID(userID string) ([]*Principal, error) {
	p.txContext.EnsureTx()
	return p.impl.GetPrincipalsByUserID(userID)
}

func (p *safeProviderImpl) GetPrincipalsByEmail(email string) ([]*Principal, error) {
	p.txContext.EnsureTx()
	return p.impl.GetPrincipalsByEmail(email)
}

func (p *safeProviderImpl) UpdatePrincipal(principal Principal) error {
	p.txContext.EnsureTx()
	return p.impl.UpdatePrincipal(principal)
}

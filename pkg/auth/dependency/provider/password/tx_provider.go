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
	loginIDsKeyWhitelist []string,
	passwordHistoryEnabled bool,
	txContext db.SafeTxContext,
) Provider {
	return &safeProviderImpl{
		impl:      newProvider(builder, executor, logger, loginIDsKeyWhitelist, passwordHistoryEnabled),
		txContext: txContext,
	}
}

func (p *safeProviderImpl) IsLoginIDValid(loginID map[string]string) bool {
	p.txContext.EnsureTx()
	return p.impl.IsLoginIDValid(loginID)
}

func (p *safeProviderImpl) IsLoginIDMatching(loginID map[string]string) bool {
	p.txContext.EnsureTx()
	return p.impl.IsLoginIDMatching(loginID)
}

func (p *safeProviderImpl) CreatePrincipalsByLoginID(authInfoID string, password string, loginID map[string]string) error {
	p.txContext.EnsureTx()
	return p.impl.CreatePrincipalsByLoginID(authInfoID, password, loginID)
}

func (p *safeProviderImpl) CreatePrincipal(principal Principal) error {
	p.txContext.EnsureTx()
	return p.impl.CreatePrincipal(principal)
}

func (p *safeProviderImpl) GetPrincipalByLoginID(loginIDKey string, loginID string, principal *Principal) (err error) {
	p.txContext.EnsureTx()
	return p.impl.GetPrincipalByLoginID(loginIDKey, loginID, principal)
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

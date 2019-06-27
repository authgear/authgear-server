package password

import (
	"github.com/sirupsen/logrus"
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
	loginIDsKeys map[string]config.LoginIDKeyConfiguration,
	allowedRealms []string,
	passwordHistoryEnabled bool,
	txContext db.SafeTxContext,
) Provider {
	return &safeProviderImpl{
		impl:      newProvider(builder, executor, logger, loginIDsKeys, allowedRealms, passwordHistoryEnabled),
		txContext: txContext,
	}
}

func (p *safeProviderImpl) IsLoginIDValid(loginIDs []LoginID) bool {
	p.txContext.EnsureTx()
	return p.impl.IsLoginIDValid(loginIDs)
}

func (p safeProviderImpl) IsRealmValid(realm string) bool {
	p.txContext.EnsureTx()
	return p.impl.IsRealmValid(realm)
}

func (p *safeProviderImpl) IsDefaultAllowedRealms() bool {
	p.txContext.EnsureTx()
	return p.impl.IsDefaultAllowedRealms()
}

func (p *safeProviderImpl) CreatePrincipalsByLoginID(authInfoID string, password string, loginIDs []LoginID, realm string) error {
	p.txContext.EnsureTx()
	return p.impl.CreatePrincipalsByLoginID(authInfoID, password, loginIDs, realm)
}

func (p *safeProviderImpl) CreatePrincipal(principal Principal) error {
	p.txContext.EnsureTx()
	return p.impl.CreatePrincipal(principal)
}

func (p *safeProviderImpl) GetPrincipalByLoginIDWithRealm(loginIDKey string, loginID string, realm string, principal *Principal) (err error) {
	p.txContext.EnsureTx()
	return p.impl.GetPrincipalByLoginIDWithRealm(loginIDKey, loginID, realm, principal)
}

func (p *safeProviderImpl) GetPrincipalsByUserID(userID string) ([]*Principal, error) {
	p.txContext.EnsureTx()
	return p.impl.GetPrincipalsByUserID(userID)
}

func (p *safeProviderImpl) GetPrincipalsByLoginID(loginIDKey string, loginID string) ([]*Principal, error) {
	p.txContext.EnsureTx()
	return p.impl.GetPrincipalsByLoginID(loginIDKey, loginID)
}

func (p *safeProviderImpl) UpdatePrincipal(principal Principal) error {
	p.txContext.EnsureTx()
	return p.impl.UpdatePrincipal(principal)
}

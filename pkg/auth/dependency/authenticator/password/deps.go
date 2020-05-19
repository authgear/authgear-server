package password

import (
	"github.com/google/wire"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/logging"
	"github.com/skygeario/skygear-server/pkg/core/time"
)

func ProvideProvider(
	sqlb db.SQLBuilder,
	sqle db.SQLExecutor,
	t time.Provider,
	lf logging.Factory,
	ph HistoryStore,
	pc *Checker,
	c *config.TenantConfiguration,
) *Provider {
	return &Provider{
		Store:           &Store{SQLBuilder: sqlb, SQLExecutor: sqle},
		Time:            t,
		Config:          c.AppConfig.Authenticator.Password,
		Logger:          lf.NewLogger("authenticator-password"),
		PasswordHistory: ph,
		PasswordChecker: pc,
	}
}

func ProvideChecker(cfg *config.TenantConfiguration, s HistoryStore) *Checker {
	policy := cfg.AppConfig.Authenticator.Password.Policy
	return &Checker{
		PwMinLength:            policy.MinLength,
		PwUppercaseRequired:    policy.UppercaseRequired,
		PwLowercaseRequired:    policy.LowercaseRequired,
		PwDigitRequired:        policy.DigitRequired,
		PwSymbolRequired:       policy.SymbolRequired,
		PwMinGuessableLevel:    policy.MinimumGuessableLevel,
		PwExcludedKeywords:     policy.ExcludedKeywords,
		PwHistorySize:          policy.HistorySize,
		PwHistoryDays:          policy.HistoryDays,
		PasswordHistoryEnabled: policy.IsPasswordHistoryEnabled(),
		PasswordHistoryStore:   s,
	}
}

func ProvideHistoryStore(
	tp time.Provider,
	sqlb db.SQLBuilder,
	sqle db.SQLExecutor,
) *HistoryStoreImpl {
	return NewHistoryStore(tp, sqlb, sqle)
}

func ProvideHousekeeper(
	phs HistoryStore,
	lf logging.Factory,
	tConfig *config.TenantConfiguration,
) *Housekeeper {
	return NewHousekeeper(
		phs,
		lf,
		tConfig.AppConfig.Authenticator.Password.Policy.HistorySize,
		tConfig.AppConfig.Authenticator.Password.Policy.HistoryDays,
		tConfig.AppConfig.Authenticator.Password.Policy.IsPasswordHistoryEnabled(),
	)
}

var DependencySet = wire.NewSet(
	ProvideProvider,
	ProvideChecker,
	ProvideHousekeeper,
	ProvideHistoryStore,
	wire.Bind(new(HistoryStore), new(*HistoryStoreImpl)),
)

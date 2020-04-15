package audit

import (
	"github.com/google/wire"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/passwordhistory"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/logging"
)

func ProvidePasswordChecker(cfg *config.TenantConfiguration, s passwordhistory.Store) *PasswordChecker {
	policy := cfg.AppConfig.Authenticator.Password.Policy
	return &PasswordChecker{
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

func ProvidePwHousekeeper(
	phs passwordhistory.Store,
	lf logging.Factory,
	tConfig *config.TenantConfiguration,
) *PwHousekeeper {
	return NewPwHousekeeper(
		phs,
		lf,
		tConfig.AppConfig.Authenticator.Password.Policy.HistorySize,
		tConfig.AppConfig.Authenticator.Password.Policy.HistoryDays,
		tConfig.AppConfig.Authenticator.Password.Policy.IsPasswordHistoryEnabled(),
	)
}

var DependencySet = wire.NewSet(
	ProvidePasswordChecker,
	ProvidePwHousekeeper,
)

package audit

import (
	"github.com/google/wire"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/passwordhistory"
	"github.com/skygeario/skygear-server/pkg/core/config"
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
		PasswordHistoryEnabled: policy.HistorySize > 0 || policy.HistoryDays > 0,
		PasswordHistoryStore:   s,
	}
}

var DependencySet = wire.NewSet(ProvidePasswordChecker)

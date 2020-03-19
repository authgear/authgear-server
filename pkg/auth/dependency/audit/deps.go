package audit

import (
	"github.com/google/wire"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/passwordhistory"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

func ProvidePasswordChecker(cfg *config.TenantConfiguration, s passwordhistory.Store) *PasswordChecker {
	return &PasswordChecker{
		PwMinLength:            cfg.AppConfig.PasswordPolicy.MinLength,
		PwUppercaseRequired:    cfg.AppConfig.PasswordPolicy.UppercaseRequired,
		PwLowercaseRequired:    cfg.AppConfig.PasswordPolicy.LowercaseRequired,
		PwDigitRequired:        cfg.AppConfig.PasswordPolicy.DigitRequired,
		PwSymbolRequired:       cfg.AppConfig.PasswordPolicy.SymbolRequired,
		PwMinGuessableLevel:    cfg.AppConfig.PasswordPolicy.MinimumGuessableLevel,
		PwExcludedKeywords:     cfg.AppConfig.PasswordPolicy.ExcludedKeywords,
		PwHistorySize:          cfg.AppConfig.PasswordPolicy.HistorySize,
		PwHistoryDays:          cfg.AppConfig.PasswordPolicy.HistoryDays,
		PasswordHistoryEnabled: cfg.AppConfig.PasswordPolicy.HistorySize > 0 || cfg.AppConfig.PasswordPolicy.HistoryDays > 0,
		PasswordHistoryStore:   s,
	}
}

var DependencySet = wire.NewSet(ProvidePasswordChecker)

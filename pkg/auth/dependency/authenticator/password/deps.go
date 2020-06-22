package password

import (
	"github.com/google/wire"

	"github.com/skygeario/skygear-server/pkg/auth/config"
)

func ProvideChecker(cfg *config.PasswordPolicyConfig, s CheckerHistoryStore) *Checker {
	return &Checker{
		PwMinLength:            cfg.MinLength,
		PwUppercaseRequired:    cfg.UppercaseRequired,
		PwLowercaseRequired:    cfg.LowercaseRequired,
		PwDigitRequired:        cfg.DigitRequired,
		PwSymbolRequired:       cfg.SymbolRequired,
		PwMinGuessableLevel:    cfg.MinimumGuessableLevel,
		PwExcludedKeywords:     cfg.ExcludedKeywords,
		PwHistorySize:          cfg.HistorySize,
		PwHistoryDays:          cfg.HistoryDays,
		PasswordHistoryEnabled: cfg.IsEnabled(),
		PasswordHistoryStore:   s,
	}
}

var DependencySet = wire.NewSet(
	NewLogger,
	wire.Struct(new(Provider), "*"),
	wire.Struct(new(Store), "*"),
	NewHousekeeperLogger,
	wire.Struct(new(Housekeeper), "*"),
	ProvideChecker,
	wire.Struct(new(HistoryStore), "*"),
	wire.Bind(new(CheckerHistoryStore), new(*HistoryStore)),
)

package password

import (
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

func ProvideChecker(cfg *config.AuthenticatorPasswordConfig, s CheckerHistoryStore) *Checker {
	return &Checker{
		PwMinLength:            *cfg.Policy.MinLength,
		PwUppercaseRequired:    cfg.Policy.UppercaseRequired,
		PwLowercaseRequired:    cfg.Policy.LowercaseRequired,
		PwDigitRequired:        cfg.Policy.DigitRequired,
		PwSymbolRequired:       cfg.Policy.SymbolRequired,
		PwMinGuessableLevel:    cfg.Policy.MinimumGuessableLevel,
		PwExcludedKeywords:     cfg.Policy.ExcludedKeywords,
		PwHistorySize:          cfg.Policy.HistorySize,
		PwHistoryDays:          cfg.Policy.HistoryDays,
		PasswordHistoryEnabled: cfg.Policy.IsEnabled(),
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

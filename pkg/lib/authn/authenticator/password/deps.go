package password

import (
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/clock"
	utilrand "github.com/authgear/authgear-server/pkg/util/rand"
)

func ProvideChecker(
	cfg *config.AuthenticatorPasswordConfig,
	featureCfg *config.AuthenticatorFeatureConfig,
	s CheckerHistoryStore,
) *Checker {
	checker := &Checker{
		PasswordHistoryStore: s,
	}
	checker.PwMinLength = *cfg.Policy.MinLength
	checker.PwUppercaseRequired = cfg.Policy.UppercaseRequired
	checker.PwLowercaseRequired = cfg.Policy.LowercaseRequired
	checker.PwAlphabetRequired = cfg.Policy.AlphabetRequired
	checker.PwDigitRequired = cfg.Policy.DigitRequired
	checker.PwSymbolRequired = cfg.Policy.SymbolRequired

	if !*featureCfg.Password.Policy.History.Disabled {
		checker.PwHistorySize = cfg.Policy.HistorySize
		checker.PwHistoryDays = cfg.Policy.HistoryDays
		checker.PasswordHistoryEnabled = cfg.Policy.IsEnabled()
	}
	if !*featureCfg.Password.Policy.MinimumGuessableLevel.Disabled {
		checker.PwMinGuessableLevel = cfg.Policy.MinimumGuessableLevel
	}
	if !*featureCfg.Password.Policy.ExcludedKeywords.Disabled {
		checker.PwExcludedKeywords = cfg.Policy.ExcludedKeywords
	}
	return checker
}

func ProvideExpiry(
	cfg *config.AuthenticatorPasswordConfig,
	c clock.Clock,
) *Expiry {
	return &Expiry{
		ForceChangeEnabled:         cfg.Expiry.ForceChange.IsEnabled(),
		ForceChangeSinceLastUpdate: cfg.Expiry.ForceChange.DurationSinceLastUpdate,
		Clock:                      c,
	}
}

func NewRandSource() Rand {
	return RandRand{utilrand.SecureRand}
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
	ProvideExpiry,
	NewRandSource,
	wire.Struct(new(Generator), "*"),
)

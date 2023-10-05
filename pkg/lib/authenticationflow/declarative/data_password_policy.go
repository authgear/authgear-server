package declarative

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type PasswordPolicyHistory struct {
	Enabled bool `json:"enabled"`
	Size    int  `json:"size,omitempty"`
	Days    int  `json:"days,omitempty"`
}

type PasswordPolicy struct {
	MinimumLength      *int                   `json:"minimum_length,omitempty"`
	UppercaseRequired  bool                   `json:"uppercase_required,omitempty"`
	LowercaseRequired  bool                   `json:"lowercase_required,omitempty"`
	AlphabetRequired   bool                   `json:"alphabet_required,omitempty"`
	DigitRequired      bool                   `json:"digit_required,omitempty"`
	SymbolRequired     bool                   `json:"symbol_required,omitempty"`
	MinimumZxcvbnScore *int                   `json:"minimum_zxcvbn_score,omitempty"`
	History            *PasswordPolicyHistory `json:"history,omitempty"`
	ExcludedKeywords   []string               `json:"excluded_keywords,omitempty"`
}

func NewPasswordPolicy(featureCfg *config.AuthenticatorFeatureConfig, c *config.PasswordPolicyConfig) *PasswordPolicy {

	policy := &PasswordPolicy{
		MinimumLength:     c.MinLength,
		UppercaseRequired: c.UppercaseRequired,
		LowercaseRequired: c.LowercaseRequired,
		AlphabetRequired:  c.AlphabetRequired,
		DigitRequired:     c.DigitRequired,
		SymbolRequired:    c.SymbolRequired,
	}

	if !*featureCfg.Password.Policy.MinimumGuessableLevel.Disabled && c.MinimumGuessableLevel > 0 {
		score := c.MinimumGuessableLevel - 1
		policy.MinimumZxcvbnScore = &score
	}

	history := &PasswordPolicyHistory{}
	if !*featureCfg.Password.Policy.History.Disabled {
		history.Enabled = c.IsEnabled()
		history.Size = c.HistorySize
		history.Days = int(c.HistoryDays)
	}
	policy.History = history

	if !*featureCfg.Password.Policy.ExcludedKeywords.Disabled && len(c.ExcludedKeywords) > 0 {
		policy.ExcludedKeywords = c.ExcludedKeywords
	}

	return policy
}

package declarative

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type PasswordPolicy struct {
	MinimumLength      *int `json:"minimum_length,omitempty"`
	UppercaseRequired  bool `json:"uppercase_required,omitempty"`
	LowercaseRequired  bool `json:"lowercase_required,omitempty"`
	AlphabetRequired   bool `json:"alphabet_required,omitempty"`
	DigitRequired      bool `json:"digit_required,omitempty"`
	SymbolRequired     bool `json:"symbol_required,omitempty"`
	MinimumZxcvbnScore *int `json:"minimum_zxcvbn_score,omitempty"`
}

func NewPasswordPolicy(c *config.PasswordPolicyConfig) *PasswordPolicy {
	policy := &PasswordPolicy{
		MinimumLength:     c.MinLength,
		UppercaseRequired: c.UppercaseRequired,
		LowercaseRequired: c.LowercaseRequired,
		AlphabetRequired:  c.AlphabetRequired,
		DigitRequired:     c.DigitRequired,
		SymbolRequired:    c.SymbolRequired,
	}
	if c.MinimumGuessableLevel > 0 {
		score := c.MinimumGuessableLevel - 1
		policy.MinimumZxcvbnScore = &score
	}
	return policy
}

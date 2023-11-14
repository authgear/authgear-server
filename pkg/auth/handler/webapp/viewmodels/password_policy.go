package viewmodels

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow/declarative"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator/password"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type PasswordPolicyViewModel struct {
	PasswordPolicies    []password.Policy
	IsNew               bool
	PasswordRulesString string
}

type PasswordPolicyViewModelOptions struct {
	IsNew bool
}

func GetDefaultPasswordPolicyViewModelOptions() *PasswordPolicyViewModelOptions {
	return &PasswordPolicyViewModelOptions{
		IsNew: false,
	}
}

// nolint: gocognit
func NewPasswordPolicyViewModel(policies []password.Policy, rules string, apiError *apierrors.APIError, opt *PasswordPolicyViewModelOptions) PasswordPolicyViewModel {
	if apiError != nil {
		if apiError.Reason == "PasswordPolicyViolated" {
			for i, policy := range policies {
				if policy.Info == nil {
					policy.Info = map[string]interface{}{}
				}

				policy.Info["x_error_is_password_policy_violated"] = true

				for _, causei := range apiError.Info["causes"].([]interface{}) {
					if cause, ok := causei.(map[string]interface{}); ok {
						if kind, ok := cause["Name"].(string); ok {
							if kind == string(policy.Name) {
								policy.Info["x_is_violated"] = true
							}
						}
					}
				}

				policies[i] = policy
			}
		}
	}
	return PasswordPolicyViewModel{PasswordPolicies: policies, IsNew: opt.IsNew, PasswordRulesString: rules}
}

func NewPasswordPolicyViewModelFromAuthflow(p *declarative.PasswordPolicy, apiError *apierrors.APIError, opt *PasswordPolicyViewModelOptions) PasswordPolicyViewModel {
	pwMinLength := 0
	if p.MinimumLength != nil {
		pwMinLength = *p.MinimumLength
	}

	pwMinGuessableLevel := 0
	if p.MinimumZxcvbnScore != nil {
		pwMinGuessableLevel = *p.MinimumZxcvbnScore + 1
	}

	checker := &password.Checker{
		PwMinLength:            pwMinLength,
		PwUppercaseRequired:    p.UppercaseRequired,
		PwLowercaseRequired:    p.LowercaseRequired,
		PwAlphabetRequired:     p.AlphabetRequired,
		PwDigitRequired:        p.DigitRequired,
		PwSymbolRequired:       p.SymbolRequired,
		PwMinGuessableLevel:    pwMinGuessableLevel,
		PwExcludedKeywords:     p.ExcludedKeywords,
		PwHistorySize:          p.History.Size,
		PwHistoryDays:          config.DurationDays(p.History.Days),
		PasswordHistoryEnabled: p.History.Enabled,
	}
	policies := checker.PasswordPolicy()
	rules := checker.PasswordRules()
	return NewPasswordPolicyViewModel(policies, rules, apiError, opt)
}

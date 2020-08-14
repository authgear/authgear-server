package viewmodels

import (
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator/password"
)

type PasswordPolicyViewModel struct {
	PasswordPolicies []password.Policy
}

func NewPasswordPolicyViewModel(policies []password.Policy, anyError interface{}) PasswordPolicyViewModel {
	if apiError := asAPIError(anyError); apiError != nil {
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
	return PasswordPolicyViewModel{PasswordPolicies: policies}
}

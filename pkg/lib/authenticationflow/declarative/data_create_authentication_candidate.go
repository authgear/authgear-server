package declarative

import (
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type CreateAuthenticationCandidate struct {
	Authentication config.AuthenticationFlowAuthentication `json:"authentication"`

	// PasswordPolicy is specific to primary_password and secondary_password.
	PasswordPolicy *PasswordPolicy `json:"password_policy,omitempty"`
}

func NewCreateAuthenticationCandidates(deps *authflow.Dependencies, step *config.AuthenticationFlowSignupFlowStep) []CreateAuthenticationCandidate {
	candidates := []CreateAuthenticationCandidate{}
	passwordPolicy := NewPasswordPolicy(deps.Config.Authenticator.Password.Policy)
	for _, b := range step.OneOf {
		switch b.Authentication {
		case config.AuthenticationFlowAuthenticationPrimaryPassword:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryPassword:
			candidates = append(candidates, CreateAuthenticationCandidate{
				Authentication: b.Authentication,
				PasswordPolicy: passwordPolicy,
			})
		case config.AuthenticationFlowAuthenticationPrimaryPasskey:
			// Cannot create passkey in this step.
			break
		case config.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail:
			fallthrough
		case config.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS:
			candidates = append(candidates, CreateAuthenticationCandidate{
				Authentication: b.Authentication,
			})
		case config.AuthenticationFlowAuthenticationSecondaryTOTP:
			candidates = append(candidates, CreateAuthenticationCandidate{
				Authentication: b.Authentication,
			})
		case config.AuthenticationFlowAuthenticationRecoveryCode:
			// Recovery code is not created in this step.
			break
		case config.AuthenticationFlowAuthenticationDeviceToken:
			// Device token is irrelevant in this step.
			break
		}
	}
	return candidates
}

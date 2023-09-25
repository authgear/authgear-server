package declarative

import (
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type CreateAuthenticationOption struct {
	Authentication config.AuthenticationFlowAuthentication `json:"authentication"`

	// PasswordPolicy is specific to primary_password and secondary_password.
	PasswordPolicy *PasswordPolicy `json:"password_policy,omitempty"`
}

func NewCreateAuthenticationOptions(deps *authflow.Dependencies, step *config.AuthenticationFlowSignupFlowStep) []CreateAuthenticationOption {
	options := []CreateAuthenticationOption{}
	passwordPolicy := NewPasswordPolicy(deps.Config.Authenticator.Password.Policy)
	for _, b := range step.OneOf {
		switch b.Authentication {
		case config.AuthenticationFlowAuthenticationPrimaryPassword:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryPassword:
			options = append(options, CreateAuthenticationOption{
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
			options = append(options, CreateAuthenticationOption{
				Authentication: b.Authentication,
			})
		case config.AuthenticationFlowAuthenticationSecondaryTOTP:
			options = append(options, CreateAuthenticationOption{
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
	return options
}

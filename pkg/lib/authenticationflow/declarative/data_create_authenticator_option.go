package declarative

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type CreateAuthenticatorOption struct {
	Authentication config.AuthenticationFlowAuthentication `json:"authentication"`

	// OTPForm is specific to OOBOTP.
	OTPForm otp.Form `json:"otp_form,omitempty"`
	// Channels is specific to OOBOTP.
	Channels []model.AuthenticatorOOBChannel `json:"channels,omitempty"`

	// PasswordPolicy is specific to primary_password and secondary_password.
	PasswordPolicy *PasswordPolicy `json:"password_policy,omitempty"`
}

func NewCreateAuthenticationOptions(deps *authflow.Dependencies, step *config.AuthenticationFlowSignupFlowStep) []CreateAuthenticatorOption {
	options := []CreateAuthenticatorOption{}
	passwordPolicy := NewPasswordPolicy(
		deps.FeatureConfig.Authenticator,
		deps.Config.Authenticator.Password.Policy,
	)
	for _, b := range step.OneOf {
		switch b.Authentication {
		case config.AuthenticationFlowAuthenticationPrimaryPassword:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryPassword:
			options = append(options, CreateAuthenticatorOption{
				Authentication: b.Authentication,
				PasswordPolicy: passwordPolicy,
			})
		case config.AuthenticationFlowAuthenticationPrimaryPasskey:
			// Cannot create passkey in this step.
			break
		case config.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail:
			purpose := otp.PurposeOOBOTP
			channels := getChannels(model.ClaimEmail, deps.Config.Authenticator.OOB)
			otpForm := getOTPForm(purpose, model.ClaimEmail, deps.Config.Authenticator.OOB.Email)
			options = append(options, CreateAuthenticatorOption{
				Authentication: b.Authentication,
				OTPForm:        otpForm,
				Channels:       channels,
			})
		case config.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS:
			purpose := otp.PurposeOOBOTP
			channels := getChannels(model.ClaimPhoneNumber, deps.Config.Authenticator.OOB)
			otpForm := getOTPForm(purpose, model.ClaimPhoneNumber, deps.Config.Authenticator.OOB.Email)
			options = append(options, CreateAuthenticatorOption{
				Authentication: b.Authentication,
				OTPForm:        otpForm,
				Channels:       channels,
			})
		case config.AuthenticationFlowAuthenticationSecondaryTOTP:
			options = append(options, CreateAuthenticatorOption{
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

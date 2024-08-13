package declarative

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/util/phone"
)

type CreateAuthenticatorOptionForOutput struct {
	Authentication config.AuthenticationFlowAuthentication `json:"authentication"`

	BotProtection *BotProtectionData `json:"bot_protection,omitempty"`
	// OTPForm is specific to OOBOTP.
	OTPForm otp.Form `json:"otp_form,omitempty"`
	// Channels is specific to OOBOTP.
	Channels []model.AuthenticatorOOBChannel `json:"channels,omitempty"`

	// PasswordPolicy is specific to primary_password and secondary_password.
	PasswordPolicy *PasswordPolicy `json:"password_policy,omitempty"`

	// Target is specific to primary_oob_otp_email, primary_oob_otp_sms, secondary_oob_otp_email, secondary_oob_otp_sms.
	Target *CreateAuthenticatorTarget `json:"target,omitempty"`
}

type CreateAuthenticatorOption struct {
	Authentication config.AuthenticationFlowAuthentication `json:"authentication"`

	BotProtection *BotProtectionData `json:"bot_protection,omitempty"`
	// OTPForm is specific to OOBOTP.
	OTPForm otp.Form `json:"otp_form,omitempty"`
	// Channels is specific to OOBOTP.
	Channels []model.AuthenticatorOOBChannel `json:"channels,omitempty"`

	// PasswordPolicy is specific to primary_password and secondary_password.
	PasswordPolicy *PasswordPolicy `json:"password_policy,omitempty"`

	// Target is specific to primary_oob_otp_email, primary_oob_otp_sms, secondary_oob_otp_email, secondary_oob_otp_sms.
	Target *CreateAuthenticatorTarget `json:"target,omitempty"`
}

type CreateAuthenticatorOptionInternal struct {
	CreateAuthenticatorOption
	UnmaskedTarget string
}

type CreateAuthenticatorTarget struct {
	MaskedDisplayName    string `json:"masked_display_name"`
	VerificationRequired bool   `json:"verification_required"`
}

func makeCreateAuthenticatorTarget(
	ctx context.Context,
	deps *authflow.Dependencies,
	flows authflow.Flows,
	oneOf config.AuthenticationFlowObjectSignupFlowOrLoginFlowOneOf,
	userID string,
) (target *CreateAuthenticatorTarget, claimValue string, isSkipped bool, err error) {
	targetStep := oneOf.GetTargetStepName()
	if targetStep != "" {
		claimValue, isSkipped, err := getCreateAuthenticatorOOBOTPTargetFromTargetStep(ctx, deps, flows, targetStep)
		if err != nil {
			return nil, "", isSkipped, err
		}
		if claimValue == "" {
			return nil, "", isSkipped, nil
		}
		claimName := getOOBAuthenticatorType(oneOf.GetAuthentication()).ToClaimName()
		verified, err := getCreateAuthenticatorOOBOTPTargetVerified(deps, userID, claimName, claimValue)
		if err != nil {
			return nil, "", isSkipped, err
		}
		masked := ""
		switch claimName {
		case model.ClaimEmail:
			masked = mail.MaskAddress(claimValue)
		case model.ClaimPhoneNumber:
			masked = phone.Mask(claimValue)
		}
		target = &CreateAuthenticatorTarget{
			MaskedDisplayName:    masked,
			VerificationRequired: !verified && oneOf.IsVerificationRequired(),
		}
	}
	return target, claimValue, isSkipped, nil
}

func NewCreateAuthenticationOptions(
	ctx context.Context,
	deps *authflow.Dependencies,
	flows authflow.Flows,
	step config.AuthenticationFlowObjectSignupFlowOrLoginFlowStep,
	userID string) ([]CreateAuthenticatorOptionInternal, error) {
	options := []CreateAuthenticatorOptionInternal{}
	passwordPolicy := NewPasswordPolicy(
		deps.FeatureConfig.Authenticator,
		deps.Config.Authenticator.Password.Policy,
	)
	oneOf := step.GetSignupFlowOrLoginFlowOneOf()
	for _, b := range oneOf {
		switch b.GetAuthentication() {
		case config.AuthenticationFlowAuthenticationPrimaryPassword:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryPassword:
			options = append(options, CreateAuthenticatorOptionInternal{
				CreateAuthenticatorOption: CreateAuthenticatorOption{
					Authentication: b.GetAuthentication(),
					PasswordPolicy: passwordPolicy,
					BotProtection:  GetBotProtectionData(b.GetBotProtectionConfig(), deps.Config.BotProtection),
				},
			})
		case config.AuthenticationFlowAuthenticationPrimaryPasskey:
			// Cannot create passkey in this step.
			break
		case config.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail:
			target, unmaskedTarget, isSkipped, err := makeCreateAuthenticatorTarget(ctx, deps, flows, b, userID)
			if err != nil {
				return nil, err
			}
			if isSkipped {
				// Skip this option, because the target step was skipped
				continue
			}
			purpose := otp.PurposeOOBOTP
			channels := getChannels(model.ClaimEmail, deps.Config.Authenticator.OOB)
			otpForm := getOTPForm(purpose, model.ClaimEmail, deps.Config.Authenticator.OOB.Email)
			options = append(options, CreateAuthenticatorOptionInternal{
				UnmaskedTarget: unmaskedTarget,
				CreateAuthenticatorOption: CreateAuthenticatorOption{
					Authentication: b.GetAuthentication(),
					OTPForm:        otpForm,
					Channels:       channels,
					Target:         target,
					BotProtection:  GetBotProtectionData(b.GetBotProtectionConfig(), deps.Config.BotProtection),
				},
			})
		case config.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS:
			target, unmaskedTarget, isSkipped, err := makeCreateAuthenticatorTarget(ctx, deps, flows, b, userID)
			if err != nil {
				return nil, err
			}
			if isSkipped {
				// Skip this option, because the target step was skipped
				continue
			}
			purpose := otp.PurposeOOBOTP
			channels := getChannels(model.ClaimPhoneNumber, deps.Config.Authenticator.OOB)
			otpForm := getOTPForm(purpose, model.ClaimPhoneNumber, deps.Config.Authenticator.OOB.Email)
			options = append(options, CreateAuthenticatorOptionInternal{
				CreateAuthenticatorOption: CreateAuthenticatorOption{
					Authentication: b.GetAuthentication(),
					OTPForm:        otpForm,
					Channels:       channels,
					Target:         target,
					BotProtection:  GetBotProtectionData(b.GetBotProtectionConfig(), deps.Config.BotProtection),
				},
				UnmaskedTarget: unmaskedTarget,
			})
		case config.AuthenticationFlowAuthenticationSecondaryTOTP:
			options = append(options, CreateAuthenticatorOptionInternal{
				CreateAuthenticatorOption: CreateAuthenticatorOption{
					Authentication: b.GetAuthentication(),
					BotProtection:  GetBotProtectionData(b.GetBotProtectionConfig(), deps.Config.BotProtection),
				},
			})
		case config.AuthenticationFlowAuthenticationRecoveryCode:
			// Recovery code is not created in this step.
			break
		case config.AuthenticationFlowAuthenticationDeviceToken:
			// Device token is irrelevant in this step.
			break
		}
	}
	return options, nil
}

func (o *CreateAuthenticatorOption) ToOutput(ctx context.Context) CreateAuthenticatorOptionForOutput {
	shdBypassBotProtection := ShouldExistingResultBypassBotProtectionRequirement(ctx)
	if shdBypassBotProtection {
		o.BotProtection = nil
	}
	return CreateAuthenticatorOptionForOutput{
		Authentication: o.Authentication,
		BotProtection:  o.BotProtection,
		OTPForm:        o.OTPForm,
		Channels:       o.Channels,
		PasswordPolicy: o.PasswordPolicy,
		Target:         o.Target,
	}
}

func (o *CreateAuthenticatorOption) isBotProtectionRequired() bool {
	if o.BotProtection == nil {
		return false
	}
	if o.BotProtection.Enabled != nil && *o.BotProtection.Enabled && o.BotProtection.Provider != nil && o.BotProtection.Provider.Type != "" {
		return true
	}

	return false
}

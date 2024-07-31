package declarative

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func GenerateSignupFlowConfig(cfg *config.AppConfig) *config.AuthenticationFlowSignupFlow {
	flow := &config.AuthenticationFlowSignupFlow{
		Name: nameGeneratedFlow,
		Steps: []*config.AuthenticationFlowSignupFlowStep{
			generateSignupFlowStepIdentify(cfg),
		},
	}

	if step, ok := generateSignupFlowStepPromptCreatePasskey(cfg); ok {
		flow.Steps = append(flow.Steps, step)
	}

	return flow
}

func generateSignupFlowStepIdentify(cfg *config.AppConfig) *config.AuthenticationFlowSignupFlowStep {
	step := &config.AuthenticationFlowSignupFlowStep{
		Name: nameStepIdentify(config.AuthenticationFlowTypeSignup),
		Type: config.AuthenticationFlowSignupFlowStepTypeIdentify,
	}

	for _, identityType := range cfg.Authentication.Identities {
		switch identityType {
		case model.IdentityTypeLoginID:
			oneOf := generateSignupFlowStepIdentifyLoginID(cfg, step.Name)
			step.OneOf = append(step.OneOf, oneOf...)
		case model.IdentityTypeOAuth:
			oneOf := generateSignupFlowStepIdentifyOAuth(cfg)
			step.OneOf = append(step.OneOf, oneOf...)
		case model.IdentityTypePasskey:
			// Cannot create paskey in this step.
			break
		}
	}

	return step
}

func generateSignupFlowStepIdentifyLoginID(cfg *config.AppConfig, stepName string) []*config.AuthenticationFlowSignupFlowOneOf {
	var output []*config.AuthenticationFlowSignupFlowOneOf

	email := false
	phone := false
	username := false
	for _, keyConfig := range cfg.Identity.LoginID.Keys {
		switch {
		case keyConfig.Type == model.LoginIDKeyTypeEmail && !email:
			email = true
			oneOf := generateSignupFlowStepIdentifyLoginIDIdentificationEmail(cfg, stepName)
			output = append(output, oneOf)

		case keyConfig.Type == model.LoginIDKeyTypePhone && !phone:
			phone = true
			oneOf := generateSignupFlowStepIdentifyLoginIDIdentificationPhone(cfg, stepName)
			output = append(output, oneOf)
		case keyConfig.Type == model.LoginIDKeyTypeUsername && !username:
			username = true
			oneOf := generateSignupFlowStepIdentifyLoginIDIdentificationUsername(cfg, stepName)
			output = append(output, oneOf)
		}
	}

	for _, oneOf := range output {
		oneOf.BotProtection = getBotProtectionRequirementsSignupOrLogin(cfg)
	}

	return output
}

func generateSignupFlowStepIdentifyLoginIDIdentificationEmail(cfg *config.AppConfig, stepName string) *config.AuthenticationFlowSignupFlowOneOf {
	oneOf := &config.AuthenticationFlowSignupFlowOneOf{
		Identification: config.AuthenticationFlowIdentificationEmail,
	}

	// Add verify step if necessary
	if *cfg.Verification.Claims.Email.Enabled && *cfg.Verification.Claims.Email.Required {
		oneOf.Steps = append(oneOf.Steps, &config.AuthenticationFlowSignupFlowStep{
			BotProtection: getBotProtectionRequirementsOOBOTPEmail(cfg),
			Type:          config.AuthenticationFlowSignupFlowStepTypeVerify,
			TargetStep:    stepName,
		})
	}

	// Add authenticate step primary if necessary
	if stepAuthenticatePrimary, ok := generateSignupFlowStepCreateAuthenticatorPrimary(cfg, oneOf.Identification); ok {
		oneOf.Steps = append(oneOf.Steps, stepAuthenticatePrimary)
	}

	// Add authenticate step secondary if necessary
	if stepAuthenticateSecondary, ok := generateSignupFlowStepCreateAuthenticatorSecondary(cfg, oneOf.Identification); ok {
		oneOf.Steps = append(oneOf.Steps, stepAuthenticateSecondary)
	}
	return oneOf
}

func generateSignupFlowStepIdentifyLoginIDIdentificationPhone(cfg *config.AppConfig, stepName string) *config.AuthenticationFlowSignupFlowOneOf {
	oneOf := &config.AuthenticationFlowSignupFlowOneOf{
		Identification: config.AuthenticationFlowIdentificationPhone,
	}

	// Add verify step if necessary
	if *cfg.Verification.Claims.PhoneNumber.Enabled && *cfg.Verification.Claims.PhoneNumber.Required {
		oneOf.Steps = append(oneOf.Steps, &config.AuthenticationFlowSignupFlowStep{
			BotProtection: getBotProtectionRequirementsOOBOTPSMS(cfg),
			Type:          config.AuthenticationFlowSignupFlowStepTypeVerify,
			TargetStep:    stepName,
		})
	}

	// Add authenticate step primary if necessary
	if stepAuthenticatePrimary, ok := generateSignupFlowStepCreateAuthenticatorPrimary(cfg, oneOf.Identification); ok {
		oneOf.Steps = append(oneOf.Steps, stepAuthenticatePrimary)
	}

	// Add authenticate step secondary if necessary
	if stepAuthenticateSecondary, ok := generateSignupFlowStepCreateAuthenticatorSecondary(cfg, oneOf.Identification); ok {
		oneOf.Steps = append(oneOf.Steps, stepAuthenticateSecondary)
	}
	return oneOf
}

func generateSignupFlowStepIdentifyLoginIDIdentificationUsername(cfg *config.AppConfig, stepName string) *config.AuthenticationFlowSignupFlowOneOf {
	oneOf := &config.AuthenticationFlowSignupFlowOneOf{
		Identification: config.AuthenticationFlowIdentificationUsername,
	}
	// Username cannot be verified.

	// Add authenticate step primary if necessary
	if stepAuthenticatePrimary, ok := generateSignupFlowStepCreateAuthenticatorPrimary(cfg, oneOf.Identification); ok {
		oneOf.Steps = append(oneOf.Steps, stepAuthenticatePrimary)
	}

	// Add authenticate step secondary if necessary
	if stepAuthenticateSecondary, ok := generateSignupFlowStepCreateAuthenticatorSecondary(cfg, oneOf.Identification); ok {
		oneOf.Steps = append(oneOf.Steps, stepAuthenticateSecondary)
	}
	return oneOf
}

func generateSignupFlowStepIdentifyOAuth(cfg *config.AppConfig) []*config.AuthenticationFlowSignupFlowOneOf {
	if len(cfg.Identity.OAuth.Providers) == 0 {
		return nil
	}

	return []*config.AuthenticationFlowSignupFlowOneOf{
		{
			Identification: config.AuthenticationFlowIdentificationOAuth,
		},
	}
}

func generateSignupFlowStepCreateAuthenticatorPrimary(cfg *config.AppConfig, identification config.AuthenticationFlowIdentification) (*config.AuthenticationFlowSignupFlowStep, bool) {
	allowed := identification.PrimaryAuthentications()

	// This identification does not require primary authentication.
	if len(allowed) == 0 {
		return nil, false
	}

	allowedMap := make(map[config.AuthenticationFlowAuthentication]struct{})
	for _, a := range allowed {
		allowedMap[a] = struct{}{}
	}

	step := &config.AuthenticationFlowSignupFlowStep{
		Name: fmt.Sprintf(nameFormatStepAuthenticatePrimary, identification),
		Type: config.AuthenticationFlowSignupFlowStepTypeCreateAuthenticator,
	}

	for _, authenticatorType := range *cfg.Authentication.PrimaryAuthenticators {
		switch authenticatorType {
		case model.AuthenticatorTypePassword:
			am := config.AuthenticationFlowAuthenticationPrimaryPassword
			if _, ok := allowedMap[am]; ok {
				oneOf := &config.AuthenticationFlowSignupFlowOneOf{
					Authentication: am,
				}
				step.OneOf = append(step.OneOf, oneOf)
			}

		case model.AuthenticatorTypeOOBEmail:
			am := config.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail
			if _, ok := allowedMap[am]; ok {
				oneOf := &config.AuthenticationFlowSignupFlowOneOf{
					Authentication: am,
					TargetStep:     nameStepIdentify(config.AuthenticationFlowTypeSignup),
					BotProtection:  getBotProtectionRequirementsOOBOTPEmail(cfg),
				}
				step.OneOf = append(step.OneOf, oneOf)
			}

		case model.AuthenticatorTypeOOBSMS:
			am := config.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS
			if _, ok := allowedMap[am]; ok {
				oneOf := &config.AuthenticationFlowSignupFlowOneOf{
					Authentication: am,
					TargetStep:     nameStepIdentify(config.AuthenticationFlowTypeSignup),
					BotProtection:  getBotProtectionRequirementsOOBOTPSMS(cfg),
				}
				step.OneOf = append(step.OneOf, oneOf)
			}
		}
	}

	return step, true
}

func generateSignupFlowStepCreateAuthenticatorSecondary(cfg *config.AppConfig, identification config.AuthenticationFlowIdentification) (*config.AuthenticationFlowSignupFlowStep, bool) {
	// Do not need this step if mode is not required.
	if cfg.Authentication.SecondaryAuthenticationMode != config.SecondaryAuthenticationModeRequired {
		return nil, false
	}

	allowed := identification.SecondaryAuthentications()

	// This identification does not require secondary authentication.
	if len(allowed) == 0 {
		return nil, false
	}

	allowedMap := make(map[config.AuthenticationFlowAuthentication]struct{})
	for _, a := range allowed {
		allowedMap[a] = struct{}{}
	}

	var recoveryCodeStep *config.AuthenticationFlowSignupFlowStep
	if !*cfg.Authentication.RecoveryCode.Disabled {
		recoveryCodeStep = &config.AuthenticationFlowSignupFlowStep{
			Type: config.AuthenticationFlowSignupFlowStepTypeViewRecoveryCode,
		}
	}

	step := &config.AuthenticationFlowSignupFlowStep{
		Name: fmt.Sprintf(nameFormatStepAuthenticateSecondary, identification),
		Type: config.AuthenticationFlowSignupFlowStepTypeCreateAuthenticator,
	}

	addOneOf := func(am config.AuthenticationFlowAuthentication, bp *config.AuthenticationFlowBotProtection) {
		if _, ok := allowedMap[am]; ok {
			oneOf := &config.AuthenticationFlowSignupFlowOneOf{
				Authentication: am,
				BotProtection:  bp,
			}
			if recoveryCodeStep != nil {
				oneOf.Steps = append(oneOf.Steps, recoveryCodeStep)
			}

			step.OneOf = append(step.OneOf, oneOf)
		}
	}

	for _, authenticatorType := range *cfg.Authentication.SecondaryAuthenticators {
		switch authenticatorType {
		case model.AuthenticatorTypePassword:
			addOneOf(config.AuthenticationFlowAuthenticationSecondaryPassword, nil)
		case model.AuthenticatorTypeOOBEmail:
			addOneOf(config.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail, getBotProtectionRequirementsOOBOTPEmail(cfg))
		case model.AuthenticatorTypeOOBSMS:
			addOneOf(config.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS, getBotProtectionRequirementsOOBOTPSMS(cfg))
		case model.AuthenticatorTypeTOTP:
			addOneOf(config.AuthenticationFlowAuthenticationSecondaryTOTP, nil)
		}
	}

	return step, true
}

func generateSignupFlowStepPromptCreatePasskey(cfg *config.AppConfig) (*config.AuthenticationFlowSignupFlowStep, bool) {
	passkeyEnabled := false
	for _, typ := range *cfg.Authentication.PrimaryAuthenticators {
		if typ == model.AuthenticatorTypePasskey {
			passkeyEnabled = true
		}
	}

	if !passkeyEnabled {
		return nil, false
	}

	return &config.AuthenticationFlowSignupFlowStep{
		Type: config.AuthenticationFlowSignupFlowStepTypePromptCreatePasskey,
	}, true
}

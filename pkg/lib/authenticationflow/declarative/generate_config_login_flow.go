package declarative

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func GenerateLoginFlowConfig(cfg *config.AppConfig) *config.AuthenticationFlowLoginFlow {
	flow := &config.AuthenticationFlowLoginFlow{
		Name: nameGeneratedFlow,
		Steps: []*config.AuthenticationFlowLoginFlowStep{
			generateLoginFlowStepIdentify(cfg),
		},
	}

	// The steps after this step contain side effects.
	// Therefore, we check account status BEFORE we perform those steps.
	flow.Steps = append(flow.Steps, generateLoginFlowStepCheckAccountStatus())

	// We always include this step because the exact condition
	// depends on the client ID, which is unknown at this point.
	flow.Steps = append(flow.Steps, generateLoginFlowStepTerminateOtherSessions())

	if step, ok := generateLoginFlowStepPromptCreatePasskey(cfg); ok {
		flow.Steps = append(flow.Steps, step)
	}

	return flow
}

func generateLoginFlowStepIdentify(cfg *config.AppConfig) *config.AuthenticationFlowLoginFlowStep {
	step := &config.AuthenticationFlowLoginFlowStep{
		Name: nameStepIdentify(config.AuthenticationFlowTypeLogin),
		Type: config.AuthenticationFlowLoginFlowStepTypeIdentify,
	}

	for _, identityType := range cfg.Authentication.Identities {
		switch identityType {
		case model.IdentityTypeLoginID:
			oneOf := generateLoginFlowStepIdentifyLoginID(cfg)
			step.OneOf = append(step.OneOf, oneOf...)
		case model.IdentityTypeOAuth:
			oneOf := generateLoginFlowStepIdentifyOAuth(cfg)
			step.OneOf = append(step.OneOf, oneOf...)
		case model.IdentityTypePasskey:
			oneOf := generateLoginFlowStepIdentifyPasskey(cfg)
			step.OneOf = append(step.OneOf, oneOf...)
		}
	}

	// Add captcha if enabled
	if hasCaptcha(cfg) {
		for _, oneOf := range step.OneOf {
			oneOf.Captcha = &config.AuthenticationFlowCaptcha{
				Mode: config.AuthenticationFlowCaptchaModeNever,
			}
		}
	}

	return step
}

func generateLoginFlowStepIdentifyLoginID(cfg *config.AppConfig) []*config.AuthenticationFlowLoginFlowOneOf {
	var output []*config.AuthenticationFlowLoginFlowOneOf

	email := false
	phone := false
	username := false
	for _, keyConfig := range cfg.Identity.LoginID.Keys {
		switch {
		case keyConfig.Type == model.LoginIDKeyTypeEmail && !email:
			email = true

			oneOf := &config.AuthenticationFlowLoginFlowOneOf{
				Identification: config.AuthenticationFlowIdentificationEmail,
			}
			output = append(output, oneOf)

			// Add authenticate step primary if necessary
			if stepAuthenticatePrimary, ok := generateLoginFlowStepAuthenticatePrimary(cfg, oneOf.Identification); ok {
				oneOf.Steps = append(oneOf.Steps, stepAuthenticatePrimary)
			}

		case keyConfig.Type == model.LoginIDKeyTypePhone && !phone:
			phone = true

			oneOf := &config.AuthenticationFlowLoginFlowOneOf{
				Identification: config.AuthenticationFlowIdentificationPhone,
			}
			output = append(output, oneOf)

			// Add authenticate step primary if necessary
			if stepAuthenticatePrimary, ok := generateLoginFlowStepAuthenticatePrimary(cfg, oneOf.Identification); ok {
				oneOf.Steps = append(oneOf.Steps, stepAuthenticatePrimary)
			}

		case keyConfig.Type == model.LoginIDKeyTypeUsername && !username:
			username = true

			oneOf := &config.AuthenticationFlowLoginFlowOneOf{
				Identification: config.AuthenticationFlowIdentificationUsername,
			}
			output = append(output, oneOf)

			// Add authenticate step primary if necessary
			if stepAuthenticatePrimary, ok := generateLoginFlowStepAuthenticatePrimary(cfg, oneOf.Identification); ok {
				oneOf.Steps = append(oneOf.Steps, stepAuthenticatePrimary)
			}
		}
	}

	return output
}

func generateLoginFlowStepIdentifyOAuth(cfg *config.AppConfig) []*config.AuthenticationFlowLoginFlowOneOf {
	if len(cfg.Identity.OAuth.Providers) == 0 {
		return nil
	}

	return []*config.AuthenticationFlowLoginFlowOneOf{
		{
			Identification: config.AuthenticationFlowIdentificationOAuth,
		},
	}
}

func generateLoginFlowStepIdentifyPasskey(cfg *config.AppConfig) []*config.AuthenticationFlowLoginFlowOneOf {
	return []*config.AuthenticationFlowLoginFlowOneOf{
		{
			Identification: config.AuthenticationFlowIdentificationPasskey,
		},
	}
}

func generateLoginFlowStepAuthenticatePrimary(cfg *config.AppConfig, identification config.AuthenticationFlowIdentification) (*config.AuthenticationFlowLoginFlowStep, bool) {
	allowed := identification.PrimaryAuthentications()

	// This identification does not require primary authentication.
	if len(allowed) == 0 {
		return nil, false
	}

	allowedMap := make(map[config.AuthenticationFlowAuthentication]struct{})
	for _, a := range allowed {
		allowedMap[a] = struct{}{}
	}

	step := &config.AuthenticationFlowLoginFlowStep{
		Name: fmt.Sprintf(nameFormatStepAuthenticatePrimary, identification),
		Type: config.AuthenticationFlowLoginFlowStepTypeAuthenticate,
	}

	for _, authenticatorType := range *cfg.Authentication.PrimaryAuthenticators {
		switch authenticatorType {
		case model.AuthenticatorTypePassword:
			am := config.AuthenticationFlowAuthenticationPrimaryPassword
			if _, ok := allowedMap[am]; ok {
				oneOf := generateLoginFlowStepAuthenticatePrimaryPassword(cfg, step.Name, identification)
				step.OneOf = append(step.OneOf, oneOf)
			}

		case model.AuthenticatorTypePasskey:
			am := config.AuthenticationFlowAuthenticationPrimaryPasskey
			if _, ok := allowedMap[am]; ok {
				oneOf := generateLoginFlowStepAuthenticatePrimaryPasskey(cfg, identification)
				step.OneOf = append(step.OneOf, oneOf)
			}

			// passkey does not require secondary authentication.

		case model.AuthenticatorTypeOOBEmail:
			am := config.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail
			if _, ok := allowedMap[am]; ok {
				oneOf := generateLoginFlowStepAuthenticatePrimaryOOBEmail(cfg, identification)
				step.OneOf = append(step.OneOf, oneOf)
			}

		case model.AuthenticatorTypeOOBSMS:
			am := config.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS
			if _, ok := allowedMap[am]; ok {
				oneOf := generateLoginFlowStepAuthenticatePrimaryOOBSMS(cfg, identification)
				step.OneOf = append(step.OneOf, oneOf)
			}
		}
	}

	return step, true
}

func generateLoginFlowStepAuthenticatePrimaryPassword(
	cfg *config.AppConfig,
	targetStep string,
	identification config.AuthenticationFlowIdentification,
) *config.AuthenticationFlowLoginFlowOneOf {
	oneOf := &config.AuthenticationFlowLoginFlowOneOf{
		Authentication: config.AuthenticationFlowAuthenticationPrimaryPassword,
	}

	// Add authenticate step secondary if necessary
	if stepAuthenticateSecondary, ok := generateLoginFlowStepAuthenticateSecondary(cfg, identification); ok {
		oneOf.Steps = append(oneOf.Steps, stepAuthenticateSecondary)
	}

	// Add change password step.
	oneOf.Steps = append(oneOf.Steps, &config.AuthenticationFlowLoginFlowStep{
		Type:       config.AuthenticationFlowLoginFlowStepTypeChangePassword,
		TargetStep: targetStep,
	})
	return oneOf
}

func generateLoginFlowStepAuthenticatePrimaryPasskey(
	cfg *config.AppConfig,
	identification config.AuthenticationFlowIdentification,
) *config.AuthenticationFlowLoginFlowOneOf {
	oneOf := &config.AuthenticationFlowLoginFlowOneOf{
		Authentication: config.AuthenticationFlowAuthenticationPrimaryPasskey,
	}
	return oneOf
}

func generateLoginFlowStepAuthenticatePrimaryOOBEmail(
	cfg *config.AppConfig,
	identification config.AuthenticationFlowIdentification,
) *config.AuthenticationFlowLoginFlowOneOf {
	oneOf := &config.AuthenticationFlowLoginFlowOneOf{
		Authentication: config.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail,
		TargetStep:     nameStepIdentify(config.AuthenticationFlowTypeLogin),
	}

	// Add authenticate step secondary if necessary
	if stepAuthenticateSecondary, ok := generateLoginFlowStepAuthenticateSecondary(cfg, identification); ok {
		oneOf.Steps = append(oneOf.Steps, stepAuthenticateSecondary)
	}
	return oneOf
}

func generateLoginFlowStepAuthenticatePrimaryOOBSMS(
	cfg *config.AppConfig,
	identification config.AuthenticationFlowIdentification,
) *config.AuthenticationFlowLoginFlowOneOf {
	am := config.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS
	oneOf := &config.AuthenticationFlowLoginFlowOneOf{
		Authentication: am,
		TargetStep:     nameStepIdentify(config.AuthenticationFlowTypeLogin),
	}

	// Add authenticate step secondary if necessary
	if stepAuthenticateSecondary, ok := generateLoginFlowStepAuthenticateSecondary(cfg, identification); ok {
		oneOf.Steps = append(oneOf.Steps, stepAuthenticateSecondary)
	}
	return oneOf
}

func generateLoginFlowStepAuthenticateSecondary(cfg *config.AppConfig, identification config.AuthenticationFlowIdentification) (*config.AuthenticationFlowLoginFlowStep, bool) {
	// This step is always present unless secondary authentication is disabled.
	if cfg.Authentication.SecondaryAuthenticationMode.IsDisabled() {
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

	// By default this step is optional.
	optional := true
	// Unless the mode is required.
	if cfg.Authentication.SecondaryAuthenticationMode == config.SecondaryAuthenticationModeRequired {
		optional = false
	}

	step := &config.AuthenticationFlowLoginFlowStep{
		Name: fmt.Sprintf(nameFormatStepAuthenticateSecondary, identification),
		Type: config.AuthenticationFlowLoginFlowStepTypeAuthenticate,
	}
	if optional {
		step.Optional = &optional
	}

	addOneOf := func(am config.AuthenticationFlowAuthentication) {
		if _, ok := allowedMap[am]; ok {
			oneOf := &config.AuthenticationFlowLoginFlowOneOf{
				Authentication: am,
			}
			step.OneOf = append(step.OneOf, oneOf)
		}
	}

	// Add the authentication allowed BOTH in the config, and this identification.
	for _, authenticatorType := range *cfg.Authentication.SecondaryAuthenticators {
		switch authenticatorType {
		case model.AuthenticatorTypePassword:
			addOneOf(config.AuthenticationFlowAuthenticationSecondaryPassword)
		case model.AuthenticatorTypeOOBEmail:
			addOneOf(config.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail)
		case model.AuthenticatorTypeOOBSMS:
			addOneOf(config.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS)
		case model.AuthenticatorTypeTOTP:
			addOneOf(config.AuthenticationFlowAuthenticationSecondaryTOTP)
		}
	}

	// The order is important here.
	// If there are other authentication, we want them to appear before recovery code.
	// If recovery code is enabled, add it to oneOf.
	if !*cfg.Authentication.RecoveryCode.Disabled {
		step.OneOf = append(step.OneOf, &config.AuthenticationFlowLoginFlowOneOf{
			Authentication: config.AuthenticationFlowAuthenticationRecoveryCode,
		})
	}

	// If device token is enabled, add it to oneOf.
	if !cfg.Authentication.DeviceToken.Disabled {
		step.OneOf = append(step.OneOf, &config.AuthenticationFlowLoginFlowOneOf{
			Authentication: config.AuthenticationFlowAuthenticationDeviceToken,
		})
	}

	return step, true
}

func generateLoginFlowStepCheckAccountStatus() *config.AuthenticationFlowLoginFlowStep {
	return &config.AuthenticationFlowLoginFlowStep{
		Type: config.AuthenticationFlowLoginFlowStepTypeCheckAccountStatus,
	}
}

func generateLoginFlowStepTerminateOtherSessions() *config.AuthenticationFlowLoginFlowStep {
	return &config.AuthenticationFlowLoginFlowStep{
		Type: config.AuthenticationFlowLoginFlowStepTypeTerminateOtherSessions,
	}
}

func generateLoginFlowStepPromptCreatePasskey(cfg *config.AppConfig) (*config.AuthenticationFlowLoginFlowStep, bool) {
	passkeyEnabled := false
	for _, typ := range *cfg.Authentication.PrimaryAuthenticators {
		if typ == model.AuthenticatorTypePasskey {
			passkeyEnabled = true
		}
	}

	if !passkeyEnabled {
		return nil, false
	}

	return &config.AuthenticationFlowLoginFlowStep{
		Type: config.AuthenticationFlowLoginFlowStepTypePromptCreatePasskey,
	}, true
}

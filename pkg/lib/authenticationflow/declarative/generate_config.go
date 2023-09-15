package declarative

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

const (
	idGeneratedFlow                   = "default"
	idStepIdentify                    = "identify"
	idFormatStepAuthenticatePrimary   = "authenticate_primary_%s"
	idFormatStepAuthenticateSecondary = "authenticate_secondary_%s"
)

func GenerateSignupFlowConfig(cfg *config.AppConfig) *config.AuthenticationFlowSignupFlow {
	flow := &config.AuthenticationFlowSignupFlow{
		ID: idGeneratedFlow,
		Steps: []*config.AuthenticationFlowSignupFlowStep{
			generateSignupFlowStepIdentify(cfg),
		},
	}

	return flow
}

func generateSignupFlowStepIdentify(cfg *config.AppConfig) *config.AuthenticationFlowSignupFlowStep {
	step := &config.AuthenticationFlowSignupFlowStep{
		ID:   idStepIdentify,
		Type: config.AuthenticationFlowSignupFlowStepTypeIdentify,
	}

	for _, identityType := range cfg.Authentication.Identities {
		switch identityType {
		case model.IdentityTypeLoginID:
			oneOf := generateSignupFlowStepIdentifyLoginID(cfg, step.ID)
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

func generateSignupFlowStepIdentifyLoginID(cfg *config.AppConfig, stepID string) []*config.AuthenticationFlowSignupFlowOneOf {
	var output []*config.AuthenticationFlowSignupFlowOneOf

	email := false
	phone := false
	username := false
	for _, keyConfig := range cfg.Identity.LoginID.Keys {
		switch {
		case keyConfig.Type == model.LoginIDKeyTypeEmail && !email:
			email = true
			oneOf := &config.AuthenticationFlowSignupFlowOneOf{
				Identification: config.AuthenticationFlowIdentificationEmail,
			}
			output = append(output, oneOf)

			// Add verify step if necessary
			if *cfg.Verification.Claims.Email.Enabled && *cfg.Verification.Claims.Email.Required {
				oneOf.Steps = append(oneOf.Steps, &config.AuthenticationFlowSignupFlowStep{

					Type:       config.AuthenticationFlowSignupFlowStepTypeVerify,
					TargetStep: stepID,
				})
			}

			// Add authenticate step primary if necessary
			if stepAuthenticatePrimary, ok := generateSignupFlowStepAuthenticatePrimary(cfg, oneOf.Identification); ok {
				oneOf.Steps = append(oneOf.Steps, stepAuthenticatePrimary)
			}

			// Add authenticate step secondary if necessary
			if stepAuthenticateSecondary, ok := generateSignupFlowStepAuthenticateSecondary(cfg, oneOf.Identification); ok {
				oneOf.Steps = append(oneOf.Steps, stepAuthenticateSecondary)
			}

		case keyConfig.Type == model.LoginIDKeyTypePhone && !phone:
			phone = true

			oneOf := &config.AuthenticationFlowSignupFlowOneOf{
				Identification: config.AuthenticationFlowIdentificationPhone,
			}
			output = append(output, oneOf)

			// Add verify step if necessary
			if *cfg.Verification.Claims.PhoneNumber.Enabled && *cfg.Verification.Claims.PhoneNumber.Required {
				oneOf.Steps = append(oneOf.Steps, &config.AuthenticationFlowSignupFlowStep{
					Type:       config.AuthenticationFlowSignupFlowStepTypeVerify,
					TargetStep: stepID,
				})
			}

			// Add authenticate step primary if necessary
			if stepAuthenticatePrimary, ok := generateSignupFlowStepAuthenticatePrimary(cfg, oneOf.Identification); ok {
				oneOf.Steps = append(oneOf.Steps, stepAuthenticatePrimary)
			}

			// Add authenticate step secondary if necessary
			if stepAuthenticateSecondary, ok := generateSignupFlowStepAuthenticateSecondary(cfg, oneOf.Identification); ok {
				oneOf.Steps = append(oneOf.Steps, stepAuthenticateSecondary)
			}

		case keyConfig.Type == model.LoginIDKeyTypeUsername && !username:
			username = true

			oneOf := &config.AuthenticationFlowSignupFlowOneOf{
				Identification: config.AuthenticationFlowIdentificationUsername,
			}

			output = append(output, oneOf)
			// Username cannot be verified.

			// Add authenticate step primary if necessary
			if stepAuthenticatePrimary, ok := generateSignupFlowStepAuthenticatePrimary(cfg, oneOf.Identification); ok {
				oneOf.Steps = append(oneOf.Steps, stepAuthenticatePrimary)
			}

			// Add authenticate step secondary if necessary
			if stepAuthenticateSecondary, ok := generateSignupFlowStepAuthenticateSecondary(cfg, oneOf.Identification); ok {
				oneOf.Steps = append(oneOf.Steps, stepAuthenticateSecondary)
			}
		}
	}

	return output
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

func generateSignupFlowStepAuthenticatePrimary(cfg *config.AppConfig, identification config.AuthenticationFlowIdentification) (*config.AuthenticationFlowSignupFlowStep, bool) {
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
		ID:   fmt.Sprintf(idFormatStepAuthenticatePrimary, identification),
		Type: config.AuthenticationFlowSignupFlowStepTypeAuthenticate,
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
					TargetStep:     idStepIdentify,
					Steps: []*config.AuthenticationFlowSignupFlowStep{
						// Must verify.
						&config.AuthenticationFlowSignupFlowStep{
							Type:       config.AuthenticationFlowSignupFlowStepTypeVerify,
							TargetStep: step.ID,
						},
					},
				}
				step.OneOf = append(step.OneOf, oneOf)
			}

		case model.AuthenticatorTypeOOBSMS:
			am := config.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS
			if _, ok := allowedMap[am]; ok {
				oneOf := &config.AuthenticationFlowSignupFlowOneOf{
					Authentication: am,
					TargetStep:     idStepIdentify,
					Steps: []*config.AuthenticationFlowSignupFlowStep{
						// Must verify.
						&config.AuthenticationFlowSignupFlowStep{
							Type:       config.AuthenticationFlowSignupFlowStepTypeVerify,
							TargetStep: step.ID,
						},
					},
				}
				step.OneOf = append(step.OneOf, oneOf)
			}
		}
	}

	return step, true
}

func generateSignupFlowStepAuthenticateSecondary(cfg *config.AppConfig, identification config.AuthenticationFlowIdentification) (*config.AuthenticationFlowSignupFlowStep, bool) {
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
			Type: config.AuthenticationFlowSignupFlowStepTypeRecoveryCode,
		}
	}

	step := &config.AuthenticationFlowSignupFlowStep{
		ID:   fmt.Sprintf(idFormatStepAuthenticateSecondary, identification),
		Type: config.AuthenticationFlowSignupFlowStepTypeAuthenticate,
	}

	addOneOf := func(am config.AuthenticationFlowAuthentication) {
		if _, ok := allowedMap[am]; ok {
			oneOf := &config.AuthenticationFlowSignupFlowOneOf{
				Authentication: am,
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
			addOneOf(config.AuthenticationFlowAuthenticationSecondaryPassword)
		case model.AuthenticatorTypeOOBEmail:
			addOneOf(config.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail)
		case model.AuthenticatorTypeOOBSMS:
			addOneOf(config.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS)
		case model.AuthenticatorTypeTOTP:
			addOneOf(config.AuthenticationFlowAuthenticationSecondaryTOTP)
		}
	}

	return step, true
}

func GenerateLoginFlowConfig(cfg *config.AppConfig) *config.AuthenticationFlowLoginFlow {
	flow := &config.AuthenticationFlowLoginFlow{
		ID: idGeneratedFlow,
		Steps: []*config.AuthenticationFlowLoginFlowStep{
			generateLoginFlowStepIdentify(cfg),
		},
	}

	// FIXME(authflow): support create passkey.

	return flow
}

func generateLoginFlowStepIdentify(cfg *config.AppConfig) *config.AuthenticationFlowLoginFlowStep {
	step := &config.AuthenticationFlowLoginFlowStep{
		ID:   idStepIdentify,
		Type: config.AuthenticationFlowLoginFlowStepTypeIdentify,
	}

	for _, identityType := range cfg.Authentication.Identities {
		switch identityType {
		case model.IdentityTypeLoginID:
			oneOf := generateLoginFlowStepIdentifyLoginID(cfg)
			step.OneOf = append(step.OneOf, oneOf...)
		case model.IdentityTypeOAuth:
			oneOf := generateLoginFlowStepIdentityOAuth(cfg)
			step.OneOf = append(step.OneOf, oneOf...)
		case model.IdentityTypePasskey:
			oneOf := generateLoginFlowStepIdentityPasskey(cfg)
			step.OneOf = append(step.OneOf, oneOf...)
			break
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

func generateLoginFlowStepIdentityOAuth(cfg *config.AppConfig) []*config.AuthenticationFlowLoginFlowOneOf {
	if len(cfg.Identity.OAuth.Providers) == 0 {
		return nil
	}

	return []*config.AuthenticationFlowLoginFlowOneOf{
		{
			Identification: config.AuthenticationFlowIdentificationOAuth,
		},
	}
}

func generateLoginFlowStepIdentityPasskey(cfg *config.AppConfig) []*config.AuthenticationFlowLoginFlowOneOf {
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
		ID:   fmt.Sprintf(idFormatStepAuthenticatePrimary, identification),
		Type: config.AuthenticationFlowLoginFlowStepTypeAuthenticate,
	}

	for _, authenticatorType := range *cfg.Authentication.PrimaryAuthenticators {
		switch authenticatorType {
		case model.AuthenticatorTypePassword:
			am := config.AuthenticationFlowAuthenticationPrimaryPassword
			if _, ok := allowedMap[am]; ok {
				oneOf := &config.AuthenticationFlowLoginFlowOneOf{
					Authentication: am,
				}
				step.OneOf = append(step.OneOf, oneOf)

				// Add change password step.
				if *cfg.Authenticator.Password.ForceChange {
					oneOf.Steps = append(oneOf.Steps, &config.AuthenticationFlowLoginFlowStep{
						Type:       config.AuthenticationFlowLoginFlowStepTypeChangePassword,
						TargetStep: step.ID,
					})
				}

				// Add authenticate step secondary if necessary
				if stepAuthenticateSecondary, ok := generateLoginFlowStepAuthenticateSecondary(cfg, identification); ok {
					oneOf.Steps = append(oneOf.Steps, stepAuthenticateSecondary)
				}
			}

		case model.AuthenticatorTypePasskey:
			am := config.AuthenticationFlowAuthenticationPrimaryPasskey
			if _, ok := allowedMap[am]; ok {
				oneOf := &config.AuthenticationFlowLoginFlowOneOf{
					Authentication: am,
				}
				step.OneOf = append(step.OneOf, oneOf)
			}

			// passkey does not require secondary authentication.

		case model.AuthenticatorTypeOOBEmail:
			am := config.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail
			if _, ok := allowedMap[am]; ok {
				oneOf := &config.AuthenticationFlowLoginFlowOneOf{
					Authentication: am,
					TargetStep:     idStepIdentify,
				}
				step.OneOf = append(step.OneOf, oneOf)

				// Add authenticate step secondary if necessary
				if stepAuthenticateSecondary, ok := generateLoginFlowStepAuthenticateSecondary(cfg, identification); ok {
					oneOf.Steps = append(oneOf.Steps, stepAuthenticateSecondary)
				}
			}

		case model.AuthenticatorTypeOOBSMS:
			am := config.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS
			if _, ok := allowedMap[am]; ok {
				oneOf := &config.AuthenticationFlowLoginFlowOneOf{
					Authentication: am,
					TargetStep:     idStepIdentify,
				}
				step.OneOf = append(step.OneOf, oneOf)

				// Add authenticate step secondary if necessary
				if stepAuthenticateSecondary, ok := generateLoginFlowStepAuthenticateSecondary(cfg, identification); ok {
					oneOf.Steps = append(oneOf.Steps, stepAuthenticateSecondary)
				}
			}
		}
	}

	return step, true
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
		ID:   fmt.Sprintf(idFormatStepAuthenticateSecondary, identification),
		Type: config.AuthenticationFlowLoginFlowStepTypeAuthenticate,
	}
	if optional {
		step.Optional = &optional
	}

	// If device token is enabled, add it to oneOf.
	if !cfg.Authentication.DeviceToken.Disabled {
		step.OneOf = append(step.OneOf, &config.AuthenticationFlowLoginFlowOneOf{
			Authentication: config.AuthenticationFlowAuthenticationDeviceToken,
		})
	}

	// If recovery code is enabled, add it to oneOf.
	if !*cfg.Authentication.RecoveryCode.Disabled {
		step.OneOf = append(step.OneOf, &config.AuthenticationFlowLoginFlowOneOf{
			Authentication: config.AuthenticationFlowAuthenticationRecoveryCode,
		})
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

	return step, true
}

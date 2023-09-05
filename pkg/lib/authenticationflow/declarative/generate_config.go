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

	email := false
	phone := false
	username := false
	for _, identityType := range cfg.Authentication.Identities {
		if identityType == model.IdentityTypeLoginID {
			for _, keyConfig := range cfg.Identity.LoginID.Keys {
				switch {
				case keyConfig.Type == model.LoginIDKeyTypeEmail && !email:
					email = true
					oneOf := &config.AuthenticationFlowSignupFlowOneOf{
						Identification: config.AuthenticationFlowIdentificationEmail,
					}
					step.OneOf = append(step.OneOf, oneOf)

					// Add verify step if necessary
					if *cfg.Verification.Claims.Email.Enabled && *cfg.Verification.Claims.Email.Required {
						oneOf.Steps = append(oneOf.Steps, &config.AuthenticationFlowSignupFlowStep{

							Type:       config.AuthenticationFlowSignupFlowStepTypeVerify,
							TargetStep: step.ID,
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
					step.OneOf = append(step.OneOf, oneOf)

					// Add verify step if necessary
					if *cfg.Verification.Claims.PhoneNumber.Enabled && *cfg.Verification.Claims.PhoneNumber.Required {
						oneOf.Steps = append(oneOf.Steps, &config.AuthenticationFlowSignupFlowStep{
							Type:       config.AuthenticationFlowSignupFlowStepTypeVerify,
							TargetStep: step.ID,
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

					step.OneOf = append(step.OneOf, oneOf)
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
		}
	}

	return step
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

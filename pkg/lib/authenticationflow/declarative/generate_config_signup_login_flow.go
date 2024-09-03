package declarative

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func GenerateSignupLoginFlowConfig(cfg *config.AppConfig) *config.AuthenticationFlowSignupLoginFlow {
	flow := &config.AuthenticationFlowSignupLoginFlow{
		Name: nameGeneratedFlow,
		Steps: []*config.AuthenticationFlowSignupLoginFlowStep{
			generateSignupLoginFlowStepIdentify(cfg),
		},
	}

	return flow
}

func generateSignupLoginFlowStepIdentify(cfg *config.AppConfig) *config.AuthenticationFlowSignupLoginFlowStep {
	step := &config.AuthenticationFlowSignupLoginFlowStep{
		Name: nameStepIdentify(config.AuthenticationFlowTypeSignupLogin),
		Type: config.AuthenticationFlowSignupLoginFlowStepTypeIdentify,
	}

	for _, identityType := range cfg.Authentication.Identities {
		switch identityType {
		case model.IdentityTypeLoginID:
			oneOf := generateSignupLoginFlowStepIdentifyLoginID(cfg)
			step.OneOf = append(step.OneOf, oneOf...)
		case model.IdentityTypeOAuth:
			oneOf := generateSignupLoginFlowStepIdentifyOAuth(cfg)
			step.OneOf = append(step.OneOf, oneOf...)
		case model.IdentityTypePasskey:
			oneOf := generateSignupLoginFlowStepIdentifyPasskey(cfg)
			step.OneOf = append(step.OneOf, oneOf...)
		case model.IdentityTypeLDAP:
			oneOf := generateSignupLoginFlowStepIdentifyLDAP(cfg)
			step.OneOf = append(step.OneOf, oneOf...)
		}
	}

	return step
}

func newSignupLoginFlowOneOf(i config.AuthenticationFlowIdentification) *config.AuthenticationFlowSignupLoginFlowOneOf {
	return &config.AuthenticationFlowSignupLoginFlowOneOf{
		Identification: i,
		SignupFlow:     nameGeneratedFlow,
		LoginFlow:      nameGeneratedFlow,
	}
}

func generateSignupLoginFlowStepIdentifyLoginID(cfg *config.AppConfig) []*config.AuthenticationFlowSignupLoginFlowOneOf {
	var output []*config.AuthenticationFlowSignupLoginFlowOneOf

	email := false
	phone := false
	username := false
	for _, keyConfig := range cfg.Identity.LoginID.Keys {
		switch {
		case keyConfig.Type == model.LoginIDKeyTypeEmail && !email:
			email = true
			oneOf := newSignupLoginFlowOneOf(config.AuthenticationFlowIdentificationEmail)
			// Add bot protection to identify step if bot_protection.requirements.oob_otp_email is required
			if bp, ok := getBotProtectionRequirementsOOBOTPEmail(cfg); ok {
				oneOf.BotProtection = bp
			}
			output = append(output, oneOf)

		case keyConfig.Type == model.LoginIDKeyTypePhone && !phone:
			phone = true
			oneOf := newSignupLoginFlowOneOf(config.AuthenticationFlowIdentificationPhone)
			// Add bot protection to identify step if bot_protection.requirements.oob_otp_sms is required
			if bp, ok := getBotProtectionRequirementsOOBOTPSMS(cfg); ok {
				oneOf.BotProtection = bp
			}
			output = append(output, oneOf)

		case keyConfig.Type == model.LoginIDKeyTypeUsername && !username:
			username = true
			output = append(output, newSignupLoginFlowOneOf(config.AuthenticationFlowIdentificationUsername))
		}
	}

	if bp, ok := getBotProtectionRequirementsSignupOrLogin(cfg); ok {
		for _, oneOf := range output {
			if oneOf.BotProtection == nil {
				oneOf.BotProtection = bp
			} else {
				oneOf.BotProtection = config.GetStrictestAuthFlowBotProtection(bp, oneOf.BotProtection)
			}
		}
	}

	return output
}

func generateSignupLoginFlowStepIdentifyOAuth(cfg *config.AppConfig) []*config.AuthenticationFlowSignupLoginFlowOneOf {
	if len(cfg.Identity.OAuth.Providers) == 0 {
		return nil
	}

	return []*config.AuthenticationFlowSignupLoginFlowOneOf{
		newSignupLoginFlowOneOf(config.AuthenticationFlowIdentificationOAuth),
	}
}

func generateSignupLoginFlowStepIdentifyPasskey(cfg *config.AppConfig) []*config.AuthenticationFlowSignupLoginFlowOneOf {
	oneOf := newSignupLoginFlowOneOf(config.AuthenticationFlowIdentificationPasskey)
	// We do not allow signup with a passkey.
	oneOf.SignupFlow = ""

	return []*config.AuthenticationFlowSignupLoginFlowOneOf{
		oneOf,
	}
}

func generateSignupLoginFlowStepIdentifyLDAP(cfg *config.AppConfig) []*config.AuthenticationFlowSignupLoginFlowOneOf {
	if len(cfg.Identity.LDAP.Servers) == 0 {
		return nil
	}

	return []*config.AuthenticationFlowSignupLoginFlowOneOf{
		newSignupLoginFlowOneOf(config.AuthenticationFlowIdentificationLDAP),
	}
}

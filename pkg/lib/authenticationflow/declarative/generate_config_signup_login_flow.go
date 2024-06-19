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
		}
	}

	// Add captcha if enabled
	if hasCaptcha(cfg) {
		for _, oneOf := range step.OneOf {
			oneOf.Captcha = &config.AuthenticationFlowCaptcha{
				Required: getBoolPtr(true),
			}
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
			output = append(output, newSignupLoginFlowOneOf(config.AuthenticationFlowIdentificationEmail))

		case keyConfig.Type == model.LoginIDKeyTypePhone && !phone:
			phone = true
			output = append(output, newSignupLoginFlowOneOf(config.AuthenticationFlowIdentificationPhone))

		case keyConfig.Type == model.LoginIDKeyTypeUsername && !username:
			username = true
			output = append(output, newSignupLoginFlowOneOf(config.AuthenticationFlowIdentificationUsername))
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

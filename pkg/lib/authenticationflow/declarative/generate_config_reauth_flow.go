package declarative

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func GenerateReauthFlowConfig(cfg *config.AppConfig) *config.AuthenticationFlowReauthFlow {
	flow := &config.AuthenticationFlowReauthFlow{
		Name: nameGeneratedFlow,
		Steps: []*config.AuthenticationFlowReauthFlowStep{
			generateReauthFlowStepIdentify(cfg),
			generateReauthFlowStepAuthenticate(cfg),
		},
	}

	return flow
}

func generateReauthFlowStepIdentify(cfg *config.AppConfig) *config.AuthenticationFlowReauthFlowStep {
	step := &config.AuthenticationFlowReauthFlowStep{
		Name: nameStepIdentify(config.AuthenticationFlowTypeReauth),
		Type: config.AuthenticationFlowReauthFlowStepTypeIdentify,
		OneOf: []*config.AuthenticationFlowReauthFlowOneOf{
			{
				Identification: model.AuthenticationFlowIdentificationIDToken,
			},
		},
	}
	return step
}

func generateReauthFlowStepAuthenticate(cfg *config.AppConfig) *config.AuthenticationFlowReauthFlowStep {
	step := &config.AuthenticationFlowReauthFlowStep{
		Name: nameStepReauthenticate,
		Type: config.AuthenticationFlowReauthFlowStepTypeAuthenticate,
	}

	addOneOf := func(authentication model.AuthenticationFlowAuthentication) {
		oneOf := &config.AuthenticationFlowReauthFlowOneOf{
			Authentication: authentication,
		}
		step.OneOf = append(step.OneOf, oneOf)
	}

	for _, authenticatorType := range *cfg.Authentication.PrimaryAuthenticators {
		switch authenticatorType {
		case model.AuthenticatorTypePassword:
			addOneOf(model.AuthenticationFlowAuthenticationPrimaryPassword)
		case model.AuthenticatorTypePasskey:
			addOneOf(model.AuthenticationFlowAuthenticationPrimaryPasskey)
		case model.AuthenticatorTypeOOBEmail:
			addOneOf(model.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail)
		case model.AuthenticatorTypeOOBSMS:
			addOneOf(model.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS)
		}
	}

	if !cfg.Authentication.SecondaryAuthenticationMode.IsDisabled() {
		for _, authenticatorType := range *cfg.Authentication.SecondaryAuthenticators {
			switch authenticatorType {
			case model.AuthenticatorTypePassword:
				addOneOf(model.AuthenticationFlowAuthenticationSecondaryPassword)
			case model.AuthenticatorTypeOOBEmail:
				addOneOf(model.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail)
			case model.AuthenticatorTypeOOBSMS:
				addOneOf(model.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS)
			case model.AuthenticatorTypeTOTP:
				addOneOf(model.AuthenticationFlowAuthenticationSecondaryTOTP)
			}
		}
	}

	return step
}

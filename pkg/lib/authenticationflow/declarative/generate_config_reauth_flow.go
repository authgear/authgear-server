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
				Identification: config.AuthenticationFlowIdentificationIDToken,
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

	addOneOf := func(authentication config.AuthenticationFlowAuthentication) {
		oneOf := &config.AuthenticationFlowReauthFlowOneOf{
			Authentication: authentication,
		}
		if hasCaptcha(cfg) {
			oneOf.Captcha = &config.AuthenticationFlowCaptcha{
				Mode: config.AuthenticationFlowCaptchaModeAlways,
				Provider: &config.AuthenticationFlowCaptchaProvider{
					Alias: cfg.Captcha.Providers[0].Alias,
				},
			}
		}
		step.OneOf = append(step.OneOf, oneOf)
	}

	for _, authenticatorType := range *cfg.Authentication.PrimaryAuthenticators {
		switch authenticatorType {
		case model.AuthenticatorTypePassword:
			addOneOf(config.AuthenticationFlowAuthenticationPrimaryPassword)
		case model.AuthenticatorTypePasskey:
			addOneOf(config.AuthenticationFlowAuthenticationPrimaryPasskey)
		case model.AuthenticatorTypeOOBEmail:
			addOneOf(config.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail)
		case model.AuthenticatorTypeOOBSMS:
			addOneOf(config.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS)
		}
	}

	if !cfg.Authentication.SecondaryAuthenticationMode.IsDisabled() {
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
	}

	return step
}

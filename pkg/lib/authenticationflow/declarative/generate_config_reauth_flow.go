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
			generateReauthFlowStepAuthenticate(cfg, nameStepReauthenticate, []config.AuthenticationFlowAuthentication{
				config.AuthenticationFlowAuthenticationPrimaryPassword,
				config.AuthenticationFlowAuthenticationPrimaryPasskey,
				config.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail,
				config.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS,
				config.AuthenticationFlowAuthenticationSecondaryPassword,
				config.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail,
				config.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS,
				config.AuthenticationFlowAuthenticationSecondaryTOTP,
			}),
			generateReauthFlowStepAuthenticateForAMRConstraints(cfg),
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

func generateReauthFlowStepAuthenticate(cfg *config.AppConfig, name string, authentications []config.AuthenticationFlowAuthentication) *config.AuthenticationFlowReauthFlowStep {
	authenticationMap := make(map[config.AuthenticationFlowAuthentication]struct{})
	for _, a := range authentications {
		authenticationMap[a] = struct{}{}
	}

	step := &config.AuthenticationFlowReauthFlowStep{
		Name: name,
		Type: config.AuthenticationFlowReauthFlowStepTypeAuthenticate,
	}

	addOneOf := func(authentication config.AuthenticationFlowAuthentication) {
		if _, ok := authenticationMap[authentication]; !ok {
			return
		}
		oneOf := &config.AuthenticationFlowReauthFlowOneOf{
			Authentication: authentication,
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

func generateReauthFlowStepAuthenticateForAMRConstraints(cfg *config.AppConfig) *config.AuthenticationFlowReauthFlowStep {
	allowed := []config.AuthenticationFlowAuthentication{
		config.AuthenticationFlowAuthenticationPrimaryPasskey,
		config.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail,
		config.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS,
		config.AuthenticationFlowAuthenticationSecondaryPassword,
		config.AuthenticationFlowAuthenticationSecondaryTOTP,
	}
	step := generateReauthFlowStepAuthenticate(
		cfg,
		nameStepReauthenticateAMRConstraints,
		allowed,
	)
	valueTrue := true
	step.ShowUntilAMRConstraintsFulfilled = &valueTrue

	return step
}

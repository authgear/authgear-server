package webapp

import (
	"fmt"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func IsIdentifyStepBotProtectionRequired(identificationType config.AuthenticationFlowIdentification, f *authflow.FlowResponse) (bool, error) {
	options := GetIdentificationOptions(f)

	for _, option := range options {
		if option.Identification == identificationType {
			return option.BotProtection.IsRequired(), nil
		}
	}

	return false, fmt.Errorf("identification type: \"%v\" not found in flow response options: [%v]", identificationType, options)
}

func IsAccountRecoveryIdentifyStepBotProtectionRequired(identificationType config.AuthenticationFlowAccountRecoveryIdentification, f *authflow.FlowResponse) (bool, error) {
	options := GetAccountRecoveryIdentificationOptions(f)
	for _, option := range options {
		if option.Identification == identificationType {
			return option.BotProtection.IsRequired(), nil
		}
	}

	return false, fmt.Errorf("identification type: \"%v\" not found in flow response options: [%v]", identificationType, options)
}

func IsAuthenticateStepBotProtectionRequired(authenticationType config.AuthenticationFlowAuthentication, f *authflow.FlowResponse) (bool, error) {
	options := GetAuthenticationOptions(f)
	for _, option := range options {
		if option.Authentication == authenticationType {
			return option.BotProtection.IsRequired(), nil
		}
	}
	return false, fmt.Errorf("authentication type: \"%v\" not found in flow response options: [%v]", authenticationType, options)
}

func IsCreateAuthenticatorStepBotProtectionRequired(authenticationType config.AuthenticationFlowAuthentication, f *authflow.FlowResponse) (bool, error) {
	options := GetCreateAuthenticatorOptions(f)
	for _, option := range options {
		if option.Authentication == authenticationType {
			return option.BotProtection.IsRequired(), nil
		}
	}
	return false, fmt.Errorf("authentication type: \"%v\" not found in flow response options: [%v]", authenticationType, options)
}

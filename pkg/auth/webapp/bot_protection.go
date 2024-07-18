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

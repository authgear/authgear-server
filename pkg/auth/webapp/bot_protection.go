package webapp

import (
	"fmt"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func IsIdentifyStepBotProtectionRequired(flowType authflow.FlowType, identificationType config.AuthenticationFlowIdentification, f *authflow.FlowResponse) (bool, error) {
	if flowType == authflow.FlowTypeAccountRecovery {
		panic("IsIdentifyStepBotProtectionRequired should not be called for account recovery flow. Please call IsAccountRecoveryIdentifyStepBotProtectionRequired instead")
	}
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

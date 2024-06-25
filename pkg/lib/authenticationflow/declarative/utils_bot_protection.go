package declarative

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/feature/botprotection"
)

func isBotProtectionRequired(authflowCfg *config.AuthenticationFlowBotProtection, appCfg *config.BotProtectionConfig) bool {
	data := GetBotProtectionData(authflowCfg, appCfg)
	return data != nil
}

func IsBotProtectionRequired(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, oneOfJSONPointer jsonpointer.T) (bool, error) {
	flowRootObject, err := findFlowRootObjectInFlow(deps, flows)
	if err != nil {
		return false, err
	}

	current, err := authflow.FlowObject(flowRootObject, oneOfJSONPointer)
	if err != nil {
		return false, err
	}

	currentBranch, ok := current.(config.AuthenticationFlowObjectBotProtectionConfigProvider)
	if !ok {
		return false, authflow.ErrInvalidJSONPointer
	}

	return isBotProtectionRequired(currentBranch.GetBotProtectionConfig(), deps.Config.BotProtection), nil
}

func NewBotProtectionVerification(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, oneOfJSONPointer jsonpointer.T) (*authflow.Node, error) {
	bpvResult := authflow.GetBotProtectionVerificationResult(ctx)

	if bpvResult == nil {
		return authflow.NewSubFlow(&IntentBotProtection{
			OneOfJSONPointer: oneOfJSONPointer,
		}), nil
	}

	switch bpvResult.Outcome {
	case authflow.BotProtectionVerificationOutcomeFailed:
		// TODO: confirm failed allow retry? assume no for now
		return nil, botprotection.ErrVerificationFailed
	case authflow.BotProtectionVerificationOutcomeServiceUnavailable:
		// allow retry
		return authflow.NewSubFlow(&IntentBotProtection{
			OneOfJSONPointer: oneOfJSONPointer,
		}), nil
	case authflow.BotProtectionVerificationOutcomeVerified:
		return nil, nil
	default:
		// unexpected
		return nil, authflow.ErrFlowNotFound
	}
}

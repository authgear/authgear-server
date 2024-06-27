package declarative

import (
	"context"
	"errors"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/feature/botprotection"
)

func init() {
	authflow.RegisterIntent(&IntentBotProtection{})
}

// IntentBotProtection (MilestoneFlowBotProtection)
//   NodeDidPerformBotProtectionVerification (MilestoneDidPerformBotProtectionVerification)

type IntentBotProtection struct {
	OneOfJSONPointer jsonpointer.T `json:"json_pointer,omitempty"`
}

var _ authflow.Intent = &IntentBotProtection{}
var _ MilestoneFlowBotProtection = &IntentBotProtection{}

func (*IntentBotProtection) Kind() string {
	return "IntentBotProtection"
}

func (*IntentBotProtection) Milestone() {}
func (*IntentBotProtection) MilestoneFlowBotProtection() {
}

func (i *IntentBotProtection) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	// The last node is NodeDidPerformBotProtectionVerification
	// So if MilestoneDidPerformBotProtectionVerification is found, this intent has finished
	_, _, ok := authflow.FindMilestoneInCurrentFlow[MilestoneDidPerformBotProtectionVerification](flows)
	if ok {
		return nil, authflow.ErrEOF
	}

	return &InputSchemaBotProtectionVerification{
		JSONPointer: i.OneOfJSONPointer,
	}, nil
}

func (i *IntentBotProtection) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	var inputBotProtectionVerification inputBotProtectionVerification
	if !authflow.AsInput(input, &inputBotProtectionVerification) {
		return nil, authflow.ErrIncompatibleInput
	}

	err := deps.BotProtection.Verify(inputBotProtectionVerification.GetBotProtectionProviderResponse())
	// only allow retry if service unavailable
	switch {
	case errors.Is(err, botprotection.ErrVerificationFailed):
		return authflow.NewNodeSimple(&NodeDidPerformBotProtectionVerification{}), authflow.ErrBotProtectionVerificationFailed
	case errors.Is(err, botprotection.ErrVerificationServiceUnavailable):
		return nil, authflow.ErrBotProtectionVerificationServiceUnavailable
	case errors.Is(err, nil):
		return authflow.NewNodeSimple(&NodeDidPerformBotProtectionVerification{}), authflow.ErrBotProtectionVerificationSuccess
	default:
		// unexpected error
		return nil, err
	}
}

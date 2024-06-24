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
//   NodeDidVerifyBotProtection (MilestoneDidVerifyBotProtection)

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
	// The last node is NodeDidVerifyBotProtection
	// So if MilestoneDidVerifyBotProtection is found, this intent has finished
	_, _, ok := authflow.FindMilestoneInCurrentFlow[MilestoneDidVerifyBotProtection](flows)
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

	err := deps.BotProtection.Verify(inputBotProtectionVerification.GetBotProtectionProviderType(), inputBotProtectionVerification.GetBotProtectionProviderResponse())

	switch {
	// TODO: confirm failed/service unavailable count as DidVerify? -- assume only fail count as DidVerify for now
	case errors.Is(err, botprotection.ErrVerificationFailed):
		return authflow.NewNodeSimple(&NodeDidVerifyBotProtection{}), authflow.ErrBotProtectionVerificationFailed
	case errors.Is(err, botprotection.ErrVerificationServiceUnavailable):
		return nil, authflow.ErrBotProtectionVerificationServiceUnavailable
	case errors.Is(err, nil):
		return authflow.NewNodeSimple(&NodeDidVerifyBotProtection{}), authflow.ErrBotProtectionVerificationSuccess
	default:
		// unexpected error
		return nil, err
	}
}

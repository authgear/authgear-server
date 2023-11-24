package declarative

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
)

func init() {
	authflow.RegisterIntent(&IntentAccountRecoveryFlowStepVerifyAccountRecoveryCode{})
}

type IntentAccountRecoveryFlowStepVerifyAccountRecoveryCode struct {
	FlowReference authflow.FlowReference `json:"flow_reference,omitempty"`
	JSONPointer   jsonpointer.T          `json:"json_pointer,omitempty"`
	StepName      string                 `json:"step_name,omitempty"`
	StartFrom     jsonpointer.T          `json:"start_from,omitempty"`
}

var _ authflow.TargetStep = &IntentAccountRecoveryFlowStepVerifyAccountRecoveryCode{}

func (i *IntentAccountRecoveryFlowStepVerifyAccountRecoveryCode) GetName() string {
	return i.StepName
}

func (i *IntentAccountRecoveryFlowStepVerifyAccountRecoveryCode) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

var _ authflow.Intent = &IntentAccountRecoveryFlowStepVerifyAccountRecoveryCode{}

func (*IntentAccountRecoveryFlowStepVerifyAccountRecoveryCode) Kind() string {
	return "IntentAccountRecoveryFlowStepVerifyAccountRecoveryCode"
}

func (i *IntentAccountRecoveryFlowStepVerifyAccountRecoveryCode) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	switch len(flows.Nearest.Nodes) {
	case 0:
		return nil, nil
	case 1:
		return &InputSchemaTakeAccountRecoveryCode{
			JSONPointer: i.JSONPointer,
		}, nil
	}

	return nil, authflow.ErrEOF
}

func (i *IntentAccountRecoveryFlowStepVerifyAccountRecoveryCode) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	switch len(flows.Nearest.Nodes) {
	case 0:
		if i.isRestored() {
			// We don't want to send the code again if this step was restored
			return authflow.NewNodeSimple(&NodeSentinel{}), nil
		}
		nextNode, err := NewNodeDoSendAccountRecoveryCode(ctx, deps, flows, i.FlowReference, i.JSONPointer, i.StartFrom)
		if err != nil {
			return nil, err
		}
		return authflow.NewNodeSimple(nextNode), nil
	case 1:
		var inputTakeAccountRecoveryCode inputTakeAccountRecoveryCode
		if authflow.AsInput(input, &inputTakeAccountRecoveryCode) {
			code := inputTakeAccountRecoveryCode.GetAccountRecoveryCode()
			milestone, ok := authflow.FindMilestone[MilestoneDoUseAccountRecoveryDestination](flows.Root)
			if ok {
				dest := milestone.MilestoneDoUseAccountRecoveryDestination()
				_, err := deps.ResetPassword.VerifyCodeWithTarget(dest.TargetLoginID, code)
				if err != nil {
					return nil, err
				}
			} else {
				// MilestoneDoUseAccountRecoveryDestination might not exist, because the flow is restored
				_, err := deps.ResetPassword.VerifyCode(code)
				if err != nil {
					return nil, err
				}
			}

			return authflow.NewNodeSimple(&NodeUseAccountRecoveryCode{Code: code}), nil
		}

	}

	return nil, authflow.ErrIncompatibleInput
}

func (i *IntentAccountRecoveryFlowStepVerifyAccountRecoveryCode) isRestored() bool {
	return isNodeRestored(i.JSONPointer, i.StartFrom)
}
